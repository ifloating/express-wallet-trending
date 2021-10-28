package models

type BnComPayload struct {
	Stream string    `json:"stream"`
	Data   BnPayload `json:"data"`
}

type BnPayload struct {
	EventType string `json:"e"`
	EventTs   int64  `json:"E"`
	Symbol    string `json:"s"`
}

type BnTicker struct {
	BnPayload
	PriceChange    string `json:"p"`
	PriceChangePct string `json:"P"`
	AvgPrice       string `json:"w"`
	LatestPrice    string `json:"c"`
	HighestPrice   string `json:"h"`
	LowestPrice    string `json:"l"`
	FFirstPrice    string `json:"o"`
	Volume         string `json:"v"`
	Gmv            string `json:"q"`
	LatestVolume   string `json:"Q"`
	StartTs        int64  `json:"O"`
	EndTs          int64  `json:"C"`
	LastTradeID    int64  `json:"L"`
}

type BnAggTrade struct {
	BnPayload
	TradePrice string `json:"p"`
	TradeNum   string `json:"q"`
	TradeTs    int64  `json:"T"`
	IsSell     bool   `json:"m"`
	DumpField  bool   `json:"M"`
}

type BnKLine struct {
	BnPayload
	KLine struct {
		StartTs      int64  `json:"t"`
		EndTs        int64  `json:"T"`
		Interval     string `json:"i"`
		HighestPrice string `json:"h"`
		LowestPrice  string `json:"l"`
		Volume       string `json:"v"`
		TradeNum     int    `json:"n"`
		IsComplete   bool   `json:"x"`
		LastTradeID  int64  `json:"L"`
		BuyVolume    string `json:"V"`
	} `json:"k"`
}
