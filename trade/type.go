package trade

import "math/big"

// Trade is a trade form Bitfinex
type Trade struct {
	ID        int64
	OriginID  int64
	TradeTime int64
	Price     big.Float
	Amount    big.Float
	Type      string
}

// Tick is the same as a candlestick (accumulated trades) (granulated datapoint)
type Tick struct {
	Open          big.Float
	Close         big.Float
	High          big.Float
	Low           big.Float
	Volume        big.Float
	LastOriginID  int64
	TickEndTime   int64
}
