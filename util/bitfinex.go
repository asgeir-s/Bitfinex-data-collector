package util

import (
	"math"
	"strconv"

	"github.com/cluda/bitfinex-api-go"
	"github.com/cluda/btcdata/trade"
)

// BitfinexTradetoTrade converts bitfinex trades to my trades
func BitfinexTradetoTrade(btxTrades []bitfinex.Trade) []trade.Trade {
	var trades []trade.Trade

	if len(btxTrades) == 0 {
		return trades
	}

	for _, bfxTrade := range btxTrades {
		price, _ := strconv.ParseFloat(bfxTrade.Price, 64)
		amount, _ := strconv.ParseFloat(bfxTrade.Amount, 64)

		trades = append(trades, trade.Trade{
			OriginID:  bfxTrade.TradeId,
			TradeTime: bfxTrade.Timestamp,
			Price:     price,
			Amount:    amount,
			Type:      bfxTrade.Type,
		})
	}
	return trades
}

func BitfinexWSTradeArrayToTrade(rawTrade []float64) trade.Trade {

	tradeType := "buy"
	if rawTrade[3] < 0 {
		tradeType = "sell"
	}

	return trade.Trade{
		OriginID:  int64(rawTrade[0]),
		TradeTime: int64(rawTrade[1]),
		Price:     rawTrade[2],
		Amount:    math.Abs(rawTrade[3]),
		Type:      tradeType,
	}
}
