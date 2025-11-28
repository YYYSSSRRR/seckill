package main

import (
	"context"
	"flag"
	"os"
	"seckill_service/internal/conf"
	"seckill_service/internal/data"

	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/config"
	"github.com/go-kratos/kratos/v2/config/file"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware/tracing"
	"github.com/go-kratos/kratos/v2/registry"
	"github.com/go-kratos/kratos/v2/transport/grpc"
	_ "go.uber.org/automaxprocs"
)

// go build -ldflags "-X main.Version=x.y.z"
var (
	// Name is the name of the compiled software.
	Name string = "seckill_service"
	// Version is the version of the compiled software.
	Version string
	// flagconf is the config flag.
	flagconf string

	id, _ = os.Hostname()
)

func init() {
	flag.StringVar(&flagconf, "conf", "./configs", "config path, eg: -conf config.yaml")
}

type App struct {
	consumer *data.KafkaConsumer
	app      *kratos.App
}

func newApp(logger log.Logger, gs *grpc.Server, r registry.Registrar, consumer *data.KafkaConsumer) *App {
	kratosApp := kratos.New(
		kratos.ID(id),
		kratos.Name(Name),
		kratos.Version(Version),
		kratos.Metadata(map[string]string{}),
		kratos.Logger(logger),
		kratos.Server(
			gs,
		),
		kratos.Registrar(r),
	)
	return &App{app: kratosApp, consumer: consumer}
}

func main() {
	flag.Parse()
	logger := log.With(log.NewStdLogger(os.Stdout),
		"ts", log.DefaultTimestamp,
		"caller", log.DefaultCaller,
		"service.id", id,
		"service.name", Name,
		"service.version", Version,
		"trace.id", tracing.TraceID(),
		"span.id", tracing.SpanID(),
	)
	c := config.New(
		config.WithSource(
			file.NewSource(flagconf),
		),
	)
	defer c.Close()

	if err := c.Load(); err != nil {
		panic(err)
	}

	var bc conf.Bootstrap
	if err := c.Scan(&bc); err != nil {
		panic(err)
	}

	app, cleanup, err := wireApp(bc.Server, bc.Data, logger, bc.Registry)
	if err != nil {
		panic(err)
	}
	defer cleanup()

	//开启一个协程消费创建订单的消息
	kafkaConsumer := app.consumer
	go func() {
		log.Info("Kafka consumer started")
		err := kafkaConsumer.ConsumeAndHandler(context.Background(), []string{"create_order_topic"})
		if err != nil {
			panic(err)
		}
	}()

	// start and wait for stop signal
	kratosApp := app.app
	if err := kratosApp.Run(); err != nil {
		panic(err)
	}
}
