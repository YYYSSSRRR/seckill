package data

import (
	"context"
	"fmt"
	"seckill_service/internal/conf"

	"log"

	"github.com/go-redis/redis/v8"
	"github.com/google/wire"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// ProviderSet is data providers.
var ProviderSet = wire.NewSet(NewData, NewGormDB,
	NewRedisClient, NewProductRepo,
	NewProductServiceClient, NewUserRepo,
	NewUserServiceClient, NewSeckillRepo,
	NewKafkaConsumer, NewMyConsumerGroupHandler,
)

// Data .
type Data struct {
	gormDB      *gorm.DB
	redisClient *redis.Client
	producer    *KafkaProducer
}

// NewData .
func NewData(gormdb *gorm.DB, client *redis.Client) (*Data, func(), error) {
	cleanup := func() {
		log.Println("closing the data resources")
	}
	kafkaProducer, err := NewKafkaProducer([]string{"localhost:9094"})
	if err != nil {
		return nil, cleanup, err
	}

	return &Data{gormDB: gormdb, redisClient: client, producer: kafkaProducer}, cleanup, nil
}

func NewGormDB(data *conf.Data) (*gorm.DB, error) {
	dsn := data.Database.Source
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	return db, nil
}

func NewRedisClient(data *conf.Data) (*redis.Client, error) {
	addr := data.Redis.Addr
	password := data.Redis.Password
	db := int(data.Redis.Db)

	client := redis.NewClient(&redis.Options{Addr: addr, Password: password, DB: db})

	//测试连接
	if err := client.Ping(context.Background()).Err(); err != nil {
		return nil, err
	}
	fmt.Printf("redis连接成功")

	return client, nil
}
