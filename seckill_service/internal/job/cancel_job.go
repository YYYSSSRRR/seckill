package job

import (
	"fmt"
	"seckill_service/internal/biz"
	"seckill_service/internal/data"
	"strconv"

	"github.com/go-kratos/kratos/v2/log"
)

type CancelJob struct {
	mq      *data.Rabbitmq
	seckill biz.SeckillRepo
}

func NewCancelJob(mq *data.Rabbitmq, seckill biz.SeckillRepo) *CancelJob {
	return &CancelJob{mq: mq, seckill: seckill}
}

func (cj *CancelJob) Start() error {
	msgs, err := cj.mq.Ch.Consume(data.OrderCancelQueue,
		"",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return err
	}

	go func() {
		for msg := range msgs {
			orderIDstr := string(msg.Body)
			fmt.Printf("orderID：%s", orderIDstr)
			orderID, err := strconv.Atoi(orderIDstr)
			if err != nil {
				log.Errorf("订单ID不合法：%s", orderIDstr)
			}

			err = cj.seckill.CancelOrder(int64(orderID))
			if err != nil {
				log.Errorf("取消订单失败：%s", err)
			}
		}
	}()
	log.Info("OrderTimeoutCancelJob started")

	return nil
}
