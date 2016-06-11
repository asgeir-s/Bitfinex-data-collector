package util

import (
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
		trades = append(trades, trade.Trade{
			OriginID:  bfxTrade.TradeId,
			TradeTime: bfxTrade.Timestamp,
			Price:     bfxTrade.Price,
			Amount:    bfxTrade.Amount,
			Type:      bfxTrade.Type,
		})
	}
	return trades
}
