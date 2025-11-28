package lock

import (
	"common"
	"context"
	"time"

	"github.com/go-redis/redis/v8"
)

func Lock(ctx context.Context, lockKey string, uuid string, expire time.Duration, r *redis.Client) (bool, error) {
	//先把lockKey加入redis中
	result := r.SetNX(ctx, lockKey, uuid, expire)
	if result.Err() != nil {
		return false, result.Err()
	}
	return true, nil
}

func Unlock(ctx context.Context, lockKey string, value string, r *redis.Client) (bool, error) {
	//释放锁需要判断是否是自己的锁+释放，所以要用lua脚本确保原子性
	scriptContent, err := common.LoadLuaScript("/Users/ysr/Documents/seckill_microservice/common/lock/unlock.lua")
	if err != nil {
		return false, err
	}
	ok, err := r.Eval(ctx, scriptContent, []string{lockKey}, value).Result()
	if err != nil {
		return false, err
	}
	//释放锁成功
	if ok == 1 {
		return true, nil
	} else {
		return false, nil
	}

}
