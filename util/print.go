package util

import "fmt"
import "github.com/cluda/btcdata/trade"

// PrintTick will print the tick
func PrintTick(tick *trade.Tick) {
	fmt.Println("{")
	fmt.Println("   Open: ", tick.Open.String())
	fmt.Println("   Close: ", tick.Close.String())
	fmt.Println("   High: ", tick.High.String())
	fmt.Println("   Low: ", tick.Low.String())
	fmt.Println("   Volume: ", tick.Volume.String())
	fmt.Println("   LastOriginID: ", tick.LastOriginID)
	fmt.Println("   TickEndTime: ", tick.TickEndTime)
	fmt.Println("},")
}

// PrintTrade will print the trade
func PrintTrade(trade *trade.Trade) {
	fmt.Println("{")
	fmt.Println("   ID: ", trade.ID)
	fmt.Println("   OriginID: ", trade.OriginID)
	fmt.Println("   Price: ", trade.Price.String())
	fmt.Println("   Amount: ", trade.Amount.String())
	fmt.Println("   TradeTime: ", trade.TradeTime)
	fmt.Println("   Type: ", trade.Type)
	fmt.Println("},")
}