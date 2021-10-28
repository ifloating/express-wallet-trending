package models

import "time"

type QueryParam struct {
	Bucket    string
	TimeSpan  string
	Symbol    string
	AggWindow string
}

type InfluxPoint struct {
	Measurement string    `json:"_measurement"`
	EventType   string    `json:"event_type"`
	Symbol      string    `json:"symbol"`
	StartTime   time.Time `json:"_start"`
	EndTime     time.Time `json:"_end"`
	Tm          time.Time `json:"_time"`
	Field       string    `json:"_field"`
	Value       float64   `json:"_value"`
}

type TickerPrice struct {
	Symbol string  `json:"symbol"`
	Price  float64 `json:"price"`
	Pct    float64 `json:"pct,omitempty"`
	Ts     int64   `json:"ts"`
}
