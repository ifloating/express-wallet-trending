package services

import (
	"time"

	"github.com/Shopify/sarama"
	log "github.com/sirupsen/logrus"
)

type KafkaProducer struct {
	asyncProducer sarama.AsyncProducer
	topic         string
}

func NewAsyncKafkaProducer(kafkaBrokers []string, topic string) (*KafkaProducer, error) {
	kafkaProducer := &KafkaProducer{}
	// new kafka producer by config
	config := sarama.NewConfig()
	config.Producer.RequiredAcks = sarama.WaitForLocal
	config.Producer.Flush.Frequency = 500 * time.Millisecond

	producer, err := sarama.NewAsyncProducer(kafkaBrokers, config)
	if err != nil {
		return nil, err
	}
	kafkaProducer.asyncProducer = producer
	kafkaProducer.topic = topic

	go func() {
		for err := range producer.Errors() {
			log.Errorln("Kafka producer error:", err)
		}
	}()

	return kafkaProducer, nil
}

func (p *KafkaProducer) AsyncSendMessage(key string, value string) {
	p.asyncProducer.Input() <- &sarama.ProducerMessage{
		Topic: p.topic,
		Key:   sarama.StringEncoder(key),
		Value: sarama.StringEncoder(value),
	}
}
