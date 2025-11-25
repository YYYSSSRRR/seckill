package data

import (
	"context"
	"fmt"
	"product_service/internal/conf"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-redis/redis/v8"
	"github.com/google/wire"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// ProviderSet is data providers.
var ProviderSet = wire.NewSet(NewData, NewGormDB, NewRedisClient, NewProductRepo)

// Data .
type Data struct {
	gormDB *gorm.DB
}

type RedisClient struct {
	client *redis.Client
}

// NewData .
func NewData(c *conf.Data, logger log.Logger, gormDB *gorm.DB) (*Data, func(), error) {

	cleanup := func() {
		log.NewHelper(logger).Info("closing the data resources")
	}
	return &Data{gormDB: gormDB}, cleanup, nil
}

func NewGormDB(c *conf.Data) (*gorm.DB, error) {
	dsn := c.Database.Source
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	return db, nil
}

func NewRedisClient(c *conf.Data) (*RedisClient, error) {
	addr := c.Redis.Addr
	password := c.Redis.Password
	db := c.Redis.Db

	//初始化redis客户端
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       int(db),
	})

	//测试连接
	if err := client.Ping(context.Background()).Err(); err != nil {
		return nil, err
	}
	fmt.Printf("redis连接成功")

	return &RedisClient{client: client}, nil
}
