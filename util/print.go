package util

import "fmt"
import "github.com/cluda/btcdata/trade"

// PrintTick will print the tick
func PrintTick(tick *trade.Tick) {
	fmt.Println("{")
	fmt.Println("   Open:", tick.Open)
	fmt.Println("   Close:", tick.Close)
	fmt.Println("   High:", tick.High)
	fmt.Println("   Low:", tick.Low)
	fmt.Println("   Volume:", tick.Volume)
	fmt.Println("   LastOriginID:", tick.LastOriginID)
	fmt.Println("   TickEndTime:", tick.TickEndTime)
	fmt.Println("},")
}

// PrintTrade will print the trade
func PrintTrade(trade *trade.Trade) {
	fmt.Println("{")
	fmt.Println("   ID:", trade.ID)
	fmt.Println("   OriginID:", trade.OriginID)
	fmt.Println("   Price:", trade.Price)
	fmt.Println("   Amount:", trade.Amount)
	fmt.Println("   TradeTime:", trade.TradeTime)
	fmt.Println("   Type:", trade.Type)
	fmt.Println("},")
}