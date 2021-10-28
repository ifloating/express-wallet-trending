package main

import (
	"express-wallet-trending/confs"
	"express-wallet-trending/services"
	"flag"
	log "github.com/sirupsen/logrus"
	"os"
	"os/signal"
	"runtime"
	"syscall"
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	configFile := flag.String("config", "", "")
	flag.Parse()
	config := &confs.ServerConf{}
	if err := config.LoadFromFile(*configFile); err != nil {
		panic(err)
	}
	kafkaProducer, err := services.NewAsyncKafkaProducer(config.KafkaConf.KafkaBrokers, config.KafkaConf.Topic)
	if err != nil {
		log.Errorf("New Async Kafka Producer error: %v", err)
		return
	}
	for _, ep := range config.BnAPIConf.Endpoints {
		scheme := config.BnAPIConf.Scheme
		host := config.BnAPIConf.BaseURL
		endpoint := ep
		go func() {
			wsService, err := services.NewWsConn(scheme, host, endpoint)
			if err != nil {
				log.Errorf("New Webscoket Connection error: %v, host: %s, endpoint: %s", err, host, endpoint)
				return
			}
			defer wsService.Close()
			if err := wsService.Handle(kafkaProducer); err != nil {
				log.Errorf("websocket handle error: %s", err.Error())
			}
		}()
	}
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL)
	for {
		select {
		case <-signals:
			return
		}
	}
}
