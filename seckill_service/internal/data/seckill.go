package data

import (
	"common"
	"context"
	"encoding/json"
	"errors"
	"log"
	v1 "proto_definitions/product/v1"
	"seckill_service/internal/biz"
	"seckill_service/internal/utils"
	"strconv"

	"github.com/streadway/amqp"
)

type OrderInfo struct {
	ProductID int64 `gorm:"column:product_id"`
	UserID    int64 `gorm:"column:user_id"`
}

type seckillRepo struct {
	data    *Data
	product biz.ProductRepo
}

func NewSeckillRepo(data *Data, product biz.ProductRepo) biz.SeckillRepo {
	return &seckillRepo{data: data, product: product}
}

func (sr *seckillRepo) SendDelayMessage(orderID int64) error {
	//发送订单，到期后进入死信队列检查订单状态，所以需要传orderID
	err := sr.data.mq.Ch.Publish(OrderDelayExchange, "order.delay", false, false,
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(strconv.FormatInt(orderID, 10)),
		},
	)
	if err != nil {
		return err
	}

	return nil
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
	res1 := sr.data.redisClient.Eval(ctx, scriptContent1, []string{decreaseKey, lockKey}, "2")
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

	//发送一个订单支付消息到延时队列中
	//fmt.Printf("发送时的orderID：%d", order.OrderID)
	//err = sr.SendDelayMessage(order.OrderID)
	//if err != nil {
	//	return false, err
	//}

	//此时应该返回给前端一个支付的链接，前端跳转第三方支付网站
	//mock支付服务就是自己手动把订单状态修改为已完成

	return true, nil
}

func (sr *seckillRepo) CancelOrder(orderID int64) error {
	ctx := context.Background()
	res := sr.data.gormDB.Table("orders").Where("id=? and pay_status=?", orderID, 0).Update("pay_status", -1)
	if res.Error != nil {
		return res.Error
	}
	log.Printf("影响行数：%d", res.RowsAffected)
	scriptRollBack, err := common.LoadLuaScript("/Users/ysr/Documents/seckill_microservice/seckill_service/internal/lua/rollbackStock.lua")
	if err != nil {
		return err
	}
	orderInfo := OrderInfo{}
	err = sr.data.gormDB.Table("orders").Where("id=?", orderID).Select("product_id", "user_id").Find(&orderInfo).Error
	if err != nil {
		return err
	}
	productInfoKey := "product:" + strconv.FormatInt(orderInfo.ProductID, 10)
	userSetKey := utils.USER_SECKILL_LOCK + strconv.FormatInt(orderInfo.ProductID, 10)
	//从redis中删
	res1 := sr.data.redisClient.Eval(ctx, scriptRollBack, []string{productInfoKey, userSetKey}, strconv.FormatInt(orderInfo.UserID, 10))
	if res1.Err() != nil {
		return res1.Err()
	}
	log.Printf("stock的值：%s", res1.Val())

	//恢复数据库中的数据
	request := &v1.DeductStockRequest{Id: orderInfo.ProductID, Num: 1}
	_, err = sr.product.AddStock(ctx, request)
	if err != nil {
		return err
	}
	return nil
}

func (sr *seckillRepo) PayOrder(orderID int64) (bool, error) {
	return true, nil
}

func (sr *seckillRepo) QueryOrder(orderID int64) (*biz.Order, error) {
	return nil, nil
}
