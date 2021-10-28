package main

import (
	"express-wallet-trending/confs"
	"express-wallet-trending/handlers"
	"flag"
	"fmt"
	"github.com/go-zoo/bone"
	log "github.com/sirupsen/logrus"
	"net/http"
	"runtime"
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	configFile := flag.String("config", "", "")
	flag.Parse()
	config := &confs.TickersSvrConf{}
	if err := config.LoadFromFile(*configFile); err != nil {
		panic(err)
	}
	tickerHandler, err := handlers.NewTickersHandler(config)
	if err != nil {
		panic(err)
	}
	mux := bone.New()
	mux.Get("/api/tickers/price/line", http.HandlerFunc(tickerHandler.Tickers))
	mux.Get("/api/tickers/price/latest", http.HandlerFunc(tickerHandler.CurrentPrice))

	address := fmt.Sprintf("%s:%d", config.Host, config.Port)
	server := &http.Server{
		Addr:    address,
		Handler: mux,
	}
	if err := server.ListenAndServe(); err != nil {
		log.Errorln(err)
	}
}
