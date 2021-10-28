package handlers

import (
	"context"
	"encoding/json"
	"express-wallet-trending/confs"
	"express-wallet-trending/models"
	"fmt"
	"github.com/go-zoo/bone"
	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api"
	log "github.com/sirupsen/logrus"
	"net/http"
	"strings"
)

type TickersHandler struct {
	Org          string
	Bucket       string
	InfluxClient influxdb2.Client
	QueryAPI     api.QueryAPI
}

func NewTickersHandler(config *confs.TickersSvrConf) (*TickersHandler, error) {
	client := influxdb2.NewClient(config.InfluxConf.SeverURL, config.InfluxConf.Token)
	queryAPI := client.QueryAPI(config.InfluxConf.Org)
	handler := &TickersHandler{
		Org:          config.InfluxConf.Org,
		Bucket:       config.InfluxConf.Bucket,
		InfluxClient: client,
		QueryAPI:     queryAPI,
	}
	return handler, nil
}

func (h *TickersHandler) Tickers(w http.ResponseWriter, r *http.Request) {
	symbols := bone.GetQuery(r, "symbol")
	intervals := bone.GetQuery(r, "interval")
	if len(symbols) == 0 || len(intervals) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"data":[], msg": "parameter error", "code": 400}`))
		return
	}
	symbol := strings.ToUpper(symbols[0])
	interval := strings.ToLower(intervals[0])
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	param := &models.QueryParam{
		Bucket: h.Bucket,
		Symbol: symbol,
	}
	switch interval {
	case "1h": // 10s, 360
		timeSpan := "-1h"
		aggWindow := "10s"
		param.TimeSpan = timeSpan
		param.AggWindow = aggWindow
	case "1d": // 4min, 360
		timeSpan := "-1d"
		aggWindow := "4m"
		param.TimeSpan = timeSpan
		param.AggWindow = aggWindow
	case "1w": // 28min, 360
		timeSpan := "-7d"
		aggWindow := "28m"
		param.TimeSpan = timeSpan
		param.AggWindow = aggWindow
	case "1m": // 2h, 360
		timeSpan := "-30d"
		aggWindow := "2h"
		param.TimeSpan = timeSpan
		param.AggWindow = aggWindow
	case "1y": // 1d, 360
		timeSpan := "-360d"
		aggWindow := "1d"
		param.TimeSpan = timeSpan
		param.AggWindow = aggWindow
	default:
		log.Errorln("parameter interval error: ", interval)
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"data":[], msg": "parameter interval error", "code": 400}`))
		return
	}
	points, err := h.QueryTickers(param)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		msg := fmt.Sprintf("{\"data\":[], \"msg\": \"%s\", \"code\": 500}", err.Error())
		_, _ = w.Write([]byte(msg))
		return
	}
	resp := struct {
		Data []models.TickerPrice `json:"data"`
		Msg  string               `json:"msg"`
		Code int                  `json:"code"`
	}{
		Data: points,
		Msg:  "OK",
		Code: 0,
	}
	res, _ := json.Marshal(resp)
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(res)
}

func (h *TickersHandler) CurrentPrice(w http.ResponseWriter, r *http.Request) {
	symbols := bone.GetQuery(r, "symbol")
	if len(symbols) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"msg": "parameter error", "code": 400, "data":{}}`))
		return
	}
	symbol := strings.ToUpper(symbols[0])
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	param := &models.QueryParam{
		Bucket:   h.Bucket,
		Symbol:   symbol,
		TimeSpan: "-30s",
	}
	point, err := h.QueryCurrentPrice(param)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		msg := fmt.Sprintf("{\"msg\": \"%s\", \"code\": 500, \"data\": {}}", err.Error())
		_, _ = w.Write([]byte(msg))
		return
	}
	resp := struct {
		Data models.TickerPrice `json:"data"`
		Msg  string             `json:"msg"`
		Code int                `json:"code"`
	}{
		Data: point,
		Msg:  "OK",
		Code: 0,
	}
	res, _ := json.Marshal(resp)
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(res)
}

func (h *TickersHandler) QueryTickers(param *models.QueryParam) (points []models.TickerPrice, err error) {
	query := fmt.Sprintf(
		`from(bucket: "%s")
                  |> range(start: %s)
                  |> filter(fn: (r) => r["_measurement"] == "24hrTicker")
                  |> filter(fn: (r) => r["_field"] == "c_price")
                  |> filter(fn: (r) => r["symbol"] == "%s") 
                  |> aggregateWindow(every: %s, fn: last, createEmpty: false)
                  |> yield(name: "last")`,
		param.Bucket,
		param.TimeSpan,
		param.Symbol,
		param.AggWindow,
	)
	result, err := h.QueryAPI.Query(context.Background(), query)
	if err == nil {
		for result.Next() {
			point := models.TickerPrice{
				Symbol: param.Symbol,
				Ts:     result.Record().Time().Unix(),
				Price:  result.Record().Value().(float64),
			}
			points = append(points, point)
		}
		if result.Err() != nil {
			log.Errorln("Query tickers error: ", result.Err().Error())
		}
	}
	return
}

func (h *TickersHandler) QueryCurrentPrice(param *models.QueryParam) (price models.TickerPrice, err error) {
	query := fmt.Sprintf(
		`from(bucket: "%s")
                  |> range(start: %s) 
                  |> filter(fn: (r) => r["_measurement"] == "24hrTicker")
                  |> filter(fn: (r) => r["_field"] == "c_price" or r["_field"] == "pct")  
                  |> filter(fn: (r) => r["symbol"] == "%s") 
                  |> last()  
                  |> pivot(rowKey: ["_time"], columnKey: ["_field"], valueColumn: "_value")`,
		param.Bucket,
		param.TimeSpan,
		param.Symbol,
	)
	result, err := h.QueryAPI.Query(context.Background(), query)
	if err == nil {
		price.Symbol = param.Symbol
		for result.Next() {
			price.Ts = result.Record().Time().Unix()
			price.Price = result.Record().ValueByKey("c_price").(float64)
			price.Pct = result.Record().ValueByKey("pct").(float64)
		}
		if result.Err() != nil {
			log.Errorln("Query current price error: ", result.Err().Error())
		}
	}
	return
}
