package data

import (
	"context"
	"fmt"

	"github.com/IBM/sarama"
	"github.com/go-kratos/kratos/v2/log"
)

// TODO topic、brokers、groupName等都放到配置文件中加载（目前先硬编码）
// 开启一个消费者需要初始化consumerGroup和handler
type KafkaConsumer struct {
	//创建一个消费者组，里面有调用接口实际消费的方法
	consumer            sarama.ConsumerGroup
	consumeGroupHandler *MyConsumerGroupHandler
}

func NewKafkaConsumer(handler *MyConsumerGroupHandler) (*KafkaConsumer, error) {
	config := sarama.NewConfig()
	config.Consumer.Offsets.Initial = sarama.OffsetNewest
	config.Consumer.Return.Errors = true
	//这里只是创建了消费者组的客户端，消费者组中一个消费者对应一个分区，是sarama.consume自动分配的
	consumerGroup, err := sarama.NewConsumerGroup([]string{"localhost:9094"}, "order_consumer", config)
	if err != nil {
		return nil, err
	}
	return &KafkaConsumer{consumer: consumerGroup, consumeGroupHandler: handler}, nil
}

func (kc *KafkaConsumer) ConsumeAndHandler(ctx context.Context, topics []string) error {
	//开始消费
	for {
		fmt.Printf("running")
		if error1 := kc.consumer.Consume(ctx, topics, kc.consumeGroupHandler); error1 != nil {
			log.Errorf("Kafka consume error:%v", error1)
		}

		if ctx.Err() != nil {
			return ctx.Err()
		}

	}
}
