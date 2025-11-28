package data

import (
	"github.com/go-kratos/kratos/v2/log"

	"github.com/IBM/sarama"
)

type KafkaProducer struct {
	producer sarama.SyncProducer
}

func NewKafkaProducer(brokers []string) (*KafkaProducer, error) {
	config := sarama.NewConfig()
	config.Producer.Return.Successes = true
	config.Producer.Return.Errors = true
	producer, err := sarama.NewSyncProducer(brokers, config)
	if err != nil {
		return nil, err
	}
	return &KafkaProducer{
		producer: producer,
	}, nil

}

func (kp *KafkaProducer) Send(msg []byte, topic string) error {
	message := &sarama.ProducerMessage{Topic: topic, Key: nil, Value: sarama.StringEncoder(msg)}
	partition, offset, err := kp.producer.SendMessage(message)
	if err != nil {
		log.Errorf("SendMessage err:", err)
		return err
	}
	log.Info("[Producer] partitionid:%d;offset:%d,value:%s\n", partition, offset, msg)
	return nil
}
