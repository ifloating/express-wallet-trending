package handlers

import (
	"encoding/json"
	"express-wallet-trending/confs"
	"express-wallet-trending/constants"
	"express-wallet-trending/models"
	"github.com/Shopify/sarama"
	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api"
	"github.com/influxdata/influxdb-client-go/v2/api/write"
	log "github.com/sirupsen/logrus"
	"strconv"
	"time"
)

type BnConsumerHandler struct {
	Org          string
	Bucket       string
	InfluxClient influxdb2.Client
	WriteAPI     api.WriteAPI
	PointsChan   chan *write.Point
}

func NewBnConsumerHandler(config *confs.BnConsumerConf) (*BnConsumerHandler, error) {
	client := influxdb2.NewClient(config.InfluxConf.SeverURL, config.InfluxConf.Token)
	writeAPI := client.WriteAPI(config.InfluxConf.Org, config.InfluxConf.Bucket)
	handler := &BnConsumerHandler{
		Org:          config.InfluxConf.Org,
		Bucket:       config.InfluxConf.Bucket,
		InfluxClient: client,
		WriteAPI:     writeAPI,
		PointsChan:   make(chan *write.Point, 1000),
	}
	go func() {
		for err := range handler.WriteAPI.Errors() {
			log.Errorln("write influx db error:", err)
		}
	}()
	go func() {
		for {
			cnt := 0
			select {
			case point := <-handler.PointsChan:
				handler.WriteAPI.WritePoint(point)
				cnt += 1
				if cnt%10 == 0 {
					handler.WriteAPI.Flush()
				}
			}
		}
	}()
	return handler, nil
}

func (h *BnConsumerHandler) TickerHandler(msg *sarama.ConsumerMessage) (err error) {
	bnTicker := &models.BnTicker{}
	if err = json.Unmarshal(msg.Value, bnTicker); err != nil {
		return
	}
	avgPrice, _ := strconv.ParseFloat(bnTicker.AvgPrice, 32)
	highestPrice, _ := strconv.ParseFloat(bnTicker.HighestPrice, 32)
	lowestPrice, _ := strconv.ParseFloat(bnTicker.LowestPrice, 32)
	latestPrice, _ := strconv.ParseFloat(bnTicker.LatestPrice, 32)
	pct, _ := strconv.ParseFloat(bnTicker.PriceChangePct, 32)
	volume, _ := strconv.ParseFloat(bnTicker.Volume, 32)
	gmv, _ := strconv.ParseFloat(bnTicker.Gmv, 32)
	point := influxdb2.NewPointWithMeasurement(bnTicker.EventType)
	point.
		AddTag("symbol", bnTicker.Symbol).
		AddTag("event_type", bnTicker.EventType).
		AddField("avg_price", avgPrice).
		AddField("h_price", highestPrice).
		AddField("l_price", lowestPrice).
		AddField("c_price", latestPrice).
		AddField("pct", pct).
		AddField("volume", volume).
		AddField("gmv", gmv).
		SetTime(time.Unix(bnTicker.EventTs/1000, 0))
	h.PointsChan <- point
	return
}

func (h *BnConsumerHandler) AggTradeHandler(msg *sarama.ConsumerMessage) (err error) {
	bnAggTrade := &models.BnAggTrade{}
	if err = json.Unmarshal(msg.Value, bnAggTrade); err != nil {
		return
	}
	tradeNum, _ := strconv.ParseFloat(bnAggTrade.TradeNum, 32)
	tradePrice, _ := strconv.ParseFloat(bnAggTrade.TradePrice, 32)
	point := influxdb2.NewPointWithMeasurement(bnAggTrade.EventType)
	point.
		AddTag("symbol", bnAggTrade.Symbol).
		AddTag("event_type", bnAggTrade.EventType).
		AddField("price", tradePrice).
		AddField("volume", tradeNum).
		AddField("ts", bnAggTrade.TradeTs).
		SetTime(time.Unix(bnAggTrade.EventTs/1000, 0))
	h.PointsChan <- point
	return
}

func (h *BnConsumerHandler) KLineHandler(msg *sarama.ConsumerMessage) (err error) {
	bnKline := &models.BnKLine{}
	if err = json.Unmarshal(msg.Value, bnKline); err != nil {
		return
	}
	highestPrice, _ := strconv.ParseFloat(bnKline.KLine.HighestPrice, 32)
	lowestPrice, _ := strconv.ParseFloat(bnKline.KLine.LowestPrice, 32)
	volume, _ := strconv.ParseFloat(bnKline.KLine.Volume, 32)
	point := influxdb2.NewPointWithMeasurement(bnKline.EventType)
	point.
		AddTag("symbol", bnKline.Symbol).
		AddTag("event_type", bnKline.EventType).
		AddTag("interval", bnKline.KLine.Interval).
		AddField("start_ts", bnKline.KLine.StartTs).
		AddField("end_ts", bnKline.KLine.EndTs).
		AddField("h_price", highestPrice).
		AddField("l_price", lowestPrice).
		AddField("volume", volume).
		SetTime(time.Unix(bnKline.EventTs/1000, 0))
	h.PointsChan <- point
	return
}

func (h *BnConsumerHandler) Handle(msg *sarama.ConsumerMessage) (err error) {
	bnPayload := &models.BnPayload{}
	if err = json.Unmarshal(msg.Value, bnPayload); err != nil {
		return
	}
	switch bnPayload.EventType {
	case constants.BNTICKER:
		return h.TickerHandler(msg)
	case constants.BNAGGTRADE:
		return h.AggTradeHandler(msg)
	case constants.BNKLINE:
		return h.KLineHandler(msg)
	}
	return
}
