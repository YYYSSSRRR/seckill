package data

import (
	"log"

	"github.com/streadway/amqp"
)

type Rabbitmq struct {
	Conn *amqp.Connection
	Ch   *amqp.Channel
}

const (
	OrderDelayQueue    = "order.delay.queue"
	OrderDelayExchange = "order.delay.exchange"
	OrderDeadExchange  = "order.dead.exchange"
	OrderCancelQueue   = "order.cancel.queue"
)

// 初始化rabbitmq
func NewRabbitmq() (*Rabbitmq, error) {
	conn, err := amqp.Dial("amqp://admin:admin@localhost:5672/")
	if err != nil {
		return nil, err
	}

	ch, err := conn.Channel()
	if err != nil {
		return nil, err
	}

	//创建延时交换机(整个队列的过期时间是10s）
	err = ch.ExchangeDeclare(OrderDelayExchange, "direct", true, false, false, false, nil)
	if err != nil {
		return nil, err
	}

	//创建死信交换机
	err = ch.ExchangeDeclare(OrderDeadExchange, "direct", true, false, false, false, nil)
	if err != nil {
		return nil, err
	}

	//创建延时队列：
	_, err = ch.QueueDeclare(
		OrderDelayQueue, true, false, false, false,
		amqp.Table{
			"x-message-ttl":             6000,
			"x-dead-letter-exchange":    OrderDeadExchange,
			"x-dead-letter-routing-key": "order.cancel",
		},
	)
	if err != nil {
		return nil, err
	}

	//创建死信队列：
	_, err = ch.QueueDeclare(OrderCancelQueue, true, false, false, false, nil)
	if err != nil {
		return nil, err
	}

	//绑定延时队列：
	err = ch.QueueBind(OrderDelayQueue, "order.delay", OrderDelayExchange, false, nil)
	if err != nil {
		return nil, err
	}

	//绑定死信队列：
	err = ch.QueueBind(OrderCancelQueue, "order.cancel", OrderDeadExchange, false, nil)
	if err != nil {
		return nil, err
	}

	log.Printf("RabbitMQ queues and exchanges initialized successfully")
	return &Rabbitmq{Conn: conn, Ch: ch}, nil
}
