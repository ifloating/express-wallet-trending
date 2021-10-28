package main

import (
	"express-wallet-trending/confs"
	"express-wallet-trending/handlers"
	"express-wallet-trending/services"
	"flag"
	"runtime"
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	configFile := flag.String("config", "", "")
	flag.Parse()
	config := &confs.BnConsumerConf{}
	if err := config.LoadFromFile(*configFile); err != nil {
		panic(err)
	}
	kafkaConsumer := &services.KafkaConsumer{
		Brokers: config.KafkaConf.KafkaBrokers,
		Topics:  config.KafkaConf.Topics,
		GroupId: config.KafkaConf.GroupId,
	}
	if err := kafkaConsumer.Init(); err != nil {
		panic(err)
	}
	defer kafkaConsumer.Close()

	handler, _ := handlers.NewBnConsumerHandler(config)

	kafkaConsumer.Consume(handler.Handle)
}
