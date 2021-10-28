package services

import (
	"github.com/Shopify/sarama"
	kafkaCluster "github.com/bsm/sarama-cluster"
	log "github.com/sirupsen/logrus"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type KafkaConsumer struct {
	Brokers  []string
	Topics   []string
	GroupId  string
	consumer *kafkaCluster.Consumer
}

func (c *KafkaConsumer) Init() (err error) {
	config := kafkaCluster.NewConfig()
	config.Consumer.Return.Errors = true
	config.Group.Return.Notifications = true
	config.Consumer.Offsets.Initial = sarama.OffsetNewest
	config.Consumer.Offsets.AutoCommit.Enable = true
	config.Consumer.Offsets.AutoCommit.Interval = 2 * time.Second
	config.Consumer.Fetch.Default = 10485760
	c.consumer, err = kafkaCluster.NewConsumer(c.Brokers, c.GroupId, c.Topics, config)
	if err != nil {
		log.Errorf("Init consumer error: %v", err)
		return
	}
	go func() {
		for err := range c.consumer.Errors() {
			log.Errorf("kafka consumer error: %v", err)
		}
	}()
	go func() {
		for ntf := range c.consumer.Notifications() {
			log.Infof("kafka consumer Notification: %v", ntf)
		}
	}()
	return
}

func (c *KafkaConsumer) Consume(handler func(msg *sarama.ConsumerMessage) error) {
	log.Infof("consumer [topics=%v, groupId=%s] beginning...", c.Topics, c.GroupId)
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL)
	for {
		select {
		case msg, more := <-c.consumer.Messages():
			if more {
				err := handler(msg)
				if err != nil {
					log.Errorf("Consumer handle message [%s], error: %v", string(msg.Value), err)
				}
				c.consumer.MarkOffset(msg, "")
			}
		case <-signals:
			return
		}
	}
}

func (c *KafkaConsumer) Close() error {
	return c.consumer.Close()
}
