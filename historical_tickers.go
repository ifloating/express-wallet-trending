package main

import (
	"encoding/json"
	"express-wallet-trending/models"
	"fmt"
	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"
)

func parseKLineApiResponse(symbol string, body []byte) (result []*models.BnTicker, err error) {
	var data [][]interface{}
	if err = json.Unmarshal(body, &data); err != nil {
		return
	}
	for _, ele := range data {
		point := &models.BnTicker{}
		point.EventType = "24hrTicker"
		point.EventTs = int64(ele[6].(float64))
		point.Symbol = symbol
		point.StartTs = int64(ele[0].(float64))
		point.EndTs = int64(ele[6].(float64))
		point.HighestPrice = ele[2].(string)
		point.LowestPrice = ele[3].(string)
		point.LatestPrice = ele[4].(string)
		point.Volume = ele[5].(string)
		point.Gmv = ele[7].(string)
		result = append(result, point)
	}
	return
}

func getKLineData(symbol string, interval string) ([]byte, error) {
	baseUrl := "https://api.binance.com/api/v3/klines"
	url := fmt.Sprintf("%s?symbol=%s&interval=%s", baseUrl, symbol, interval)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	return body, err
}

func main() {
	org := "expwallet_trending"
	bucket := "expwallet_trending"
	dbUrl := "https://auto-n25-04-03-8086.ams.op-mobile.opera.com"
	token := "suoouUloipXlnSlRSirS7cEVpToMuzyIgOw0bwI0xFZZIYbmkvJj1TtgTtuA1Ec2Dh8DhxEmoUW1_iYuIdBmog=="
	client := influxdb2.NewClient(dbUrl, token)
	writeAPI := client.WriteAPI(org, bucket)
	symbols := []string{"BTCUSDT", "ETHUSDT"}
	intervals := []string{"30m", "1h", "12h", "1d"}
	for _, symbol := range symbols {
		for _, interval := range intervals {
			data, err := getKLineData(symbol, interval)
			if err != nil {
				log.Errorln(err)
				return
			}
			tickers, err := parseKLineApiResponse(symbol, data)
			if err != nil {
				log.Errorln(err)
				return
			}
			for _, ticker := range tickers {
				if ticker.EventTs < 1634814000000 { // ticker timestamp before 2021-10-21T11:00:00Z UTC
					highestPrice, _ := strconv.ParseFloat(ticker.HighestPrice, 32)
					lowestPrice, _ := strconv.ParseFloat(ticker.LowestPrice, 32)
					latestPrice, _ := strconv.ParseFloat(ticker.LatestPrice, 32)
					volume, _ := strconv.ParseFloat(ticker.Volume, 32)
					gmv, _ := strconv.ParseFloat(ticker.Gmv, 32)
					point := influxdb2.NewPointWithMeasurement(ticker.EventType)
					point.
						AddTag("symbol", ticker.Symbol).
						AddTag("event_type", ticker.EventType).
						AddField("h_price", highestPrice).
						AddField("l_price", lowestPrice).
						AddField("c_price", latestPrice).
						AddField("volume", volume).
						AddField("gmv", gmv).
						SetTime(time.Unix(ticker.EventTs/1000, 0))
					writeAPI.WritePoint(point)
				}
			}
			writeAPI.Flush()
		}
	}
	client.Close()
}
