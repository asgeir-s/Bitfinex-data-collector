package trade

// Trade is a trade form Bitfinex
type Trade struct {
	ID        int64
	OriginID  int64
	TradeTime int64
	Price     float64
	Amount    float64
	Type      string
}

// Tick is the same as a candlestick (accumulated trades) (granulated datapoint)
type Tick struct {
	Open          float64
	Close         float64
	High          float64
	Low           float64
	Volume        float64
	LastOriginID  int64
	TickEndTime   int64
}
