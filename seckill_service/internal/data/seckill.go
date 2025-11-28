package data

import (
	"common"
	"context"
	"encoding/json"
	"errors"
	"seckill_service/internal/biz"
	"seckill_service/internal/utils"
	"strconv"
)

type seckillRepo struct {
	data *Data
}

func NewSeckillRepo(data *Data) biz.SeckillRepo {
	return &seckillRepo{data: data}
}

func (sr *seckillRepo) CreateOrder(order *biz.Order) (bool, error) {
	ctx := context.Background()
	//在redis中查看用户是否已下过单，加入集合成功就是下过单，失败就是没下过单
	//TODO 如果中间失败需要把集合删除
	scriptContent, err := common.LoadLuaScript("/Users/ysr/Documents/seckill_microservice/seckill_service/internal/lua/userRepeatedOrder.lua")
	if err != nil {
		return false, err
	}
	lockKey := utils.USER_SECKILL_LOCK + strconv.FormatInt(order.ProductID, 10)
	res := sr.data.redisClient.Eval(ctx, scriptContent, []string{lockKey}, strconv.FormatInt(order.UserID, 10))
	if res.Err() != nil {
		return false, res.Err()
	}
	//加锁失败，直接返回
	if res.Val() == int64(0) {
		return false, errors.New("已经下过单啦")
	}

	//否则继续判断是否库存充足，在redis中扣减库存，扣减库存后用kafka发送消息，同步数据库的数据
	scriptContent1, err := common.LoadLuaScript("/Users/ysr/Documents/seckill_microservice/seckill_service/internal/lua/decreaseStock.lua")
	if err != nil {
		return false, err
	}
	decreaseKey := "product:" + strconv.FormatInt(order.ProductID, 10)
	res1 := sr.data.redisClient.Eval(ctx, scriptContent1, []string{decreaseKey, lockKey}, strconv.FormatInt(order.UserID, 10))
	if res1.Err() != nil {
		return false, res1.Err()
	}
	if res1.Val() == 0 {
		return false, errors.New("库存不足")
	}

	//返回值为1代表下单成功，
	//消息队列把订单写到数据库中，数据库中再扣减库存
	msg, err := json.Marshal(order)
	if err != nil {
		return false, nil
	}
	err = sr.data.producer.Send(msg, "create_order_topic")
	if err != nil {
		return false, err
	}
	return true, nil
}

func (sr *seckillRepo) PayOrder(orderID int64) (bool, error) {
	return true, nil
}

func (sr *seckillRepo) QueryOrder(orderID int64) (*biz.Order, error) {
	return nil, nil
}
