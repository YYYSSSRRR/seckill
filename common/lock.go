package common

import (
	"context"
	"time"

	"github.com/go-redis/redis/v8"
	"go.mongodb.org/mongo-driver/x/mongo/driver/uuid"
)

type DistributedLock interface {
	Lock(ctx context.Context, lockKey string, expire time.Duration) (bool, error)
	Unlock(ctx context.Context, lockKey string, value string) (bool, error)
}

type distributedLock struct {
	redisClient *redis.Client
}

func NewDistributedLock(redisClient *redis.Client) DistributedLock {
	return &distributedLock{redisClient: redisClient}
}

func (l *distributedLock) Lock(ctx context.Context, lockKey string, expire time.Duration) (bool, error) {
	lockUUID, err := uuid.New()
	if err != nil {
		return false, nil
	}
	//先把lockKey加入redis中
	result := l.redisClient.SetNX(ctx, lockKey, lockUUID, expire)
	if result.Err() != nil {
		return false, result.Err()
	}
	return true, nil
}

func (l *distributedLock) Unlock(ctx context.Context, lockKey string, value string) (bool, error) {
	return true, nil
}
