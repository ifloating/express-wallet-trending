package test

import (
	"encoding/json"
	"express-wallet-trending/confs"
	"express-wallet-trending/handlers"
	"express-wallet-trending/models"
	"fmt"
	"testing"
)

func TestInfluxDBQuery(t *testing.T) {
	config := &confs.TickersSvrConf{
		Host: "127.0.0.1",
		Port: 8080,
		InfluxConf: confs.InfluxDBConf{
			Org:      "expwallet_trending",
			Bucket:   "expwallet_trending",
			Token:    "suoouUloipXlnSlRSirS7cEVpToMuzyIgOw0bwI0xFZZIYbmkvJj1TtgTtuA1Ec2Dh8DhxEmoUW1_iYuIdBmog==",
			SeverURL: "https://auto-n25-04-03-8086.ams.op-mobile.opera.com",
		},
	}
	handler, err := handlers.NewTickersHandler(config)
	if err != nil {
		t.Error(err)
		return
	}
	param := &models.QueryParam{
		Bucket:    config.InfluxConf.Bucket,
		Symbol:    "BTCUSDT",
		TimeSpan:  "-1h",
		AggWindow: "10s",
	}
	result, err := handler.QueryTickers(param)
	if err != nil {
		t.Error(err)
		return
	}
	res, _ := json.Marshal(result)
	fmt.Println(string(res))
}
