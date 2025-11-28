package data

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	v1 "proto_definitions/product/v1"
	"seckill_service/internal/biz"

	"github.com/go-kratos/kratos/v2/log"

	"github.com/IBM/sarama"
)

// 要实现sarama.ConsumerGroup接口
type MyConsumerGroupHandler struct {
	data          *Data
	productClient biz.ProductRepo
}

func NewMyConsumerGroupHandler(data *Data, productClient biz.ProductRepo) *MyConsumerGroupHandler {
	return &MyConsumerGroupHandler{
		data:          data,
		productClient: productClient,
	}
}

func (h *MyConsumerGroupHandler) Setup(_ sarama.ConsumerGroupSession) error {
	fmt.Printf("Set up")
	return nil
}

func (h *MyConsumerGroupHandler) Cleanup(_ sarama.ConsumerGroupSession) error {
	fmt.Printf("Cleanup")
	return nil
}

// 消费逻辑在这里写
func (h *MyConsumerGroupHandler) ConsumeClaim(sess sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for msg := range claim.Messages() {
		fmt.Printf("收到消息")
		var order biz.Order
		err := json.Unmarshal(msg.Value, &order)
		if err != nil {
			return errors.New("消息类型不对，解析json错误")
		}

		//实现真正逻辑
		res := h.data.gormDB.Create(&order)
		if res.Error != nil {
			log.Errorf("sync order error:%v", err)
			return res.Error
		}
		fmt.Printf("创建订单成功")
		//数据库更新库存
		deductStockReq := &v1.DeductStockRequest{Num: 1, Id: order.ProductID}
		_, err = h.productClient.DeductStock(context.Background(), deductStockReq)
		if err != nil {
			return err
		}
		fmt.Printf("数据库缓存更新成功")

		//消费完成，标记消息消费完成
		sess.MarkMessage(msg, "")
	}
	return nil
}
