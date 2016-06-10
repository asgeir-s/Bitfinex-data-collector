package trade_test

import (
	"fmt"
	"math/big"
	"testing"
	"github.com/cluda/btcdata/util"

	"github.com/cluda/btcdata/trade"
)

var testTrades = []trade.Trade{
	{
		ID:        0,
		OriginID:  1002,
		TradeTime: 100000000,
		Price:     *big.NewFloat(500),
		Amount:    *big.NewFloat(1),
		Type:      "buy",
	}, {
		ID:        1,
		OriginID:  1004,
		TradeTime: 100000100,
		Price:     *big.NewFloat(600),
		Amount:    *big.NewFloat(1),
		Type:      "buy",
	},
	{
		ID:        2,
		OriginID:  1014,
		TradeTime: 100001100,
		Price:     *big.NewFloat(400),
		Amount:    *big.NewFloat(1),
		Type:      "sell",
	},
	{
		ID:        3,
		OriginID:  1015,
		TradeTime: 100100100,
		Price:     *big.NewFloat(500),
		Amount:    *big.NewFloat(1),
		Type:      "sell",
	},
	{
		ID:        4,
		OriginID:  1016,
		TradeTime: 101000100,
		Price:     *big.NewFloat(600),
		Amount:    *big.NewFloat(1),
		Type:      "sell",
	},
	{
		ID:        5,
		OriginID:  1030,
		TradeTime: 102000100,
		Price:     *big.NewFloat(800),
		Amount:    *big.NewFloat(1),
		Type:      "buy",
	},
}

var testTick = trade.Tick{
	Open:         *big.NewFloat(500),
	Close:        *big.NewFloat(400),
	High:         *big.NewFloat(600),
	Low:          *big.NewFloat(400),
	Volume:       *big.NewFloat(3),
	LastOriginID: 1014,
	TickEndTime:  100003000,
}

var granularityObjectTrade = trade.InitializeGranularityFromTrade(testTrades[0], "tableTestName", 3000)

func TestInitializeGranularityFromTrade(t *testing.T) {

	fmt.Println(granularityObjectTrade)
	if granularityObjectTrade.CurrentTick.Close.Cmp(&testTrades[0].Price) != 0 {
		t.Errorf("fail")
	}
	if granularityObjectTrade.CurrentTick.Open.Cmp(&testTrades[0].Price) != 0 {
		t.Errorf("fail")
	}
	if granularityObjectTrade.CurrentTick.Low.Cmp(&testTrades[0].Price) != 0 {
		t.Errorf("fail")
	}
	if granularityObjectTrade.CurrentTick.High.Cmp(&testTrades[0].Price) != 0 {
		t.Errorf("fail")
	}
	if granularityObjectTrade.CurrentTick.Volume.Cmp(&testTrades[0].Amount) != 0 {
		t.Errorf("fail")
	}
	if granularityObjectTrade.CurrentTick.LastOriginID != testTrades[0].OriginID {
		t.Errorf("fail")
	}
	if granularityObjectTrade.CurrentTick.TickEndTime != testTrades[0].TradeTime+3000 {
		t.Errorf("fail")
	}
}

func TestGranulateStartFromTrade(t *testing.T) {
	// add one trade
	ticks := trade.Granulate(testTrades[1], &granularityObjectTrade)
	if len(ticks) != 0 {
		t.Errorf("fail")
	}
	if granularityObjectTrade.CurrentTick.Close.Cmp(&testTrades[1].Price) != 0 {
		t.Errorf("fail: CurrentTick.Close: " + granularityObjectTrade.CurrentTick.Close.String() + 
		" and testTrades[1].Price: " + testTrades[1].Price.String() + " should be equal")
	}
	if granularityObjectTrade.CurrentTick.High.Cmp(&testTrades[1].Price) != 0 {
		t.Errorf("fail")
	}

	if granularityObjectTrade.CurrentTick.Open.Cmp(&testTrades[0].Price) != 0 {
		t.Errorf("fail")
	}

	if granularityObjectTrade.CurrentTick.Low.Cmp(&testTrades[0].Price) != 0 {
		t.Errorf("fail")
	}

	if granularityObjectTrade.CurrentTick.Volume.Cmp(big.NewFloat(2)) != 0 {
		t.Errorf("fail")
	}

	if granularityObjectTrade.CurrentTick.LastOriginID != testTrades[1].OriginID {
		t.Errorf("fail")
	}

	if granularityObjectTrade.CurrentTick.TickEndTime != testTrades[0].TradeTime+3000 {
		t.Errorf("fail")
	}

	// add secound trade
	ticks = trade.Granulate(testTrades[2], &granularityObjectTrade)
	if len(ticks) != 0 {
		t.Errorf("fail")
	}
	if granularityObjectTrade.CurrentTick.Close.Cmp(&testTrades[2].Price) != 0 {
		t.Errorf("fail")
	}

	if granularityObjectTrade.CurrentTick.High.Cmp(&testTrades[1].Price) != 0 {
		t.Errorf("fail")
	}

	if granularityObjectTrade.CurrentTick.Open.Cmp(&testTrades[0].Price) != 0 {
		t.Errorf("fail")
	}

	if granularityObjectTrade.CurrentTick.Low.Cmp(&testTrades[2].Price) != 0 {
		t.Errorf("fail")
	}

	if granularityObjectTrade.CurrentTick.Volume.Cmp(big.NewFloat(3)) != 0 {
		t.Errorf("fail")
	}

	if granularityObjectTrade.CurrentTick.LastOriginID != testTrades[2].OriginID {
		t.Errorf("fail")
	}

	if granularityObjectTrade.CurrentTick.TickEndTime != testTrades[0].TradeTime+3000 {
		t.Errorf("fail")
	}

	// add therd trade
	ticks = trade.Granulate(testTrades[3], &granularityObjectTrade)
	if len(ticks) == 0 {
		t.Errorf("fail")
	}

	if granularityObjectTrade.CurrentTick.Close.Cmp(&testTrades[3].Price) != 0 {
		t.Errorf("fail")
	}

	if granularityObjectTrade.CurrentTick.High.Cmp(&testTrades[3].Price) != 0 {
		t.Errorf("high should be %s but was %s", testTrades[3].Price.String(), granularityObjectTrade.CurrentTick.High.String())
	}

	if granularityObjectTrade.CurrentTick.Open.Cmp(&testTrades[3].Price) != 0 {
		t.Errorf("fail")
	}

	if granularityObjectTrade.CurrentTick.Low.Cmp(&testTrades[3].Price) != 0 {
		t.Errorf("fail")
	}

	if granularityObjectTrade.CurrentTick.Volume.Cmp(&testTrades[3].Amount) != 0 {
		t.Errorf("fail")
	}

	if granularityObjectTrade.CurrentTick.LastOriginID != testTrades[3].OriginID {
		t.Errorf("fail")
	}

	if granularityObjectTrade.CurrentTick.TickEndTime != ticks[32].TickEndTime+3000 {
		t.Errorf("fail")
	}

	if ticks[20].Open.Cmp(&testTrades[2].Price) != 0 {
		t.Errorf("fail")
	}

	if ticks[20].Close.Cmp(&testTrades[2].Price) != 0 {
		t.Errorf("fail")
	}
	if ticks[20].High.Cmp(&testTrades[2].Price) != 0 {
		t.Errorf("fail")
	}

	if ticks[20].Low.Cmp(&testTrades[2].Price) != 0 {
		t.Errorf("fail")
	}

	if ticks[20].Volume.Cmp(big.NewFloat(0)) != 0 {
		t.Errorf("fail")
	}

	if ticks[20].LastOriginID != testTrades[2].OriginID {
		t.Errorf("fail")
	}

	fmt.Println("new ticks", len(ticks))
	fmt.Println("ticks[0]:")
	util.PrintTick(&ticks[0])
	fmt.Println("ticks[10]:")
	util.PrintTick(&ticks[10])
	fmt.Println("ticks[20]:")
	util.PrintTick(&ticks[20])
	fmt.Println("ticks[32]:")
	util.PrintTick(&ticks[32])
	fmt.Println("CurrentTick")
	util.PrintTick(&granularityObjectTrade.CurrentTick)

	// add forth trade
	ticks = trade.Granulate(testTrades[4], &granularityObjectTrade)
	if len(ticks) == 0 {
		t.Errorf("fail")
	}

	if granularityObjectTrade.CurrentTick.Close.Cmp(&testTrades[4].Price) != 0 {
		t.Errorf("fail")
	}

	if granularityObjectTrade.CurrentTick.High.Cmp(&testTrades[4].Price) != 0 {
		t.Errorf("high should be %s but was %s", testTrades[4].Price.String(), granularityObjectTrade.CurrentTick.High.String())
	}

	if granularityObjectTrade.CurrentTick.Open.Cmp(&testTrades[4].Price) != 0 {
		t.Errorf("fail")
	}

	if granularityObjectTrade.CurrentTick.Low.Cmp(&testTrades[4].Price) != 0 {
		t.Errorf("fail")
	}

	if granularityObjectTrade.CurrentTick.Volume.Cmp(&testTrades[4].Amount) != 0 {
		t.Errorf("fail")
	}

	if granularityObjectTrade.CurrentTick.LastOriginID != testTrades[4].OriginID {
		t.Errorf("fail")
	}

	if ticks[20].Open.Cmp(&testTrades[3].Price) != 0 {
		t.Errorf("fail")
	}

	if ticks[20].Close.Cmp(&testTrades[3].Price) != 0 {
		t.Errorf("fail")
	}
	if ticks[20].High.Cmp(&testTrades[3].Price) != 0 {
		t.Errorf("fail")
	}

	if ticks[20].Low.Cmp(&testTrades[3].Price) != 0 {
		t.Errorf("fail")
	}

	if ticks[20].Volume.Cmp(big.NewFloat(0)) != 0 {
		t.Errorf("fail")
	}

	if ticks[20].LastOriginID != testTrades[3].OriginID {
		t.Errorf("fail")
	}

	fmt.Println("new ticks", len(ticks))
	fmt.Println("ticks[0]:")
	util.PrintTick(&ticks[0])
	fmt.Println("ticks[10]:")
	util.PrintTick(&ticks[10])
	fmt.Println("ticks[20]:")
	util.PrintTick(&ticks[20])
	fmt.Println("ticks[299]:")
	util.PrintTick(&ticks[299])
	fmt.Println("CurrentTick")
	util.PrintTick(&granularityObjectTrade.CurrentTick)
}

func TestGranulateStartFromTick(t *testing.T) {
  var granularityObjectTick = trade.InitializeGranularityFromTick(testTick, "tableTestName", 3000)

	ticks := trade.Granulate(testTrades[3], &granularityObjectTick)
	if len(ticks) == 0 {
		t.Errorf("fail")
	}

	if granularityObjectTick.CurrentTick.Close.Cmp(&testTrades[3].Price) != 0 {
		t.Errorf("fail")
	}

	if granularityObjectTick.CurrentTick.High.Cmp(&testTrades[3].Price) != 0 {
		t.Errorf("high should be %s but was %s", testTrades[3].Price.String(), granularityObjectTick.CurrentTick.High.String())
	}

	if granularityObjectTick.CurrentTick.Open.Cmp(&testTrades[3].Price) != 0 {
		t.Errorf("fail")
	}

	if granularityObjectTick.CurrentTick.Low.Cmp(&testTrades[3].Price) != 0 {
		t.Errorf("fail")
	}

	if granularityObjectTick.CurrentTick.Volume.Cmp(&testTrades[3].Amount) != 0 {
		t.Errorf("fail")
	}

	if granularityObjectTick.CurrentTick.LastOriginID != testTrades[3].OriginID {
		t.Errorf("fail")
	}

	if granularityObjectTick.CurrentTick.TickEndTime != ticks[31].TickEndTime+3000 {
		t.Errorf("fail")
	}

	if ticks[20].Open.Cmp(&testTrades[2].Price) != 0 {
		t.Errorf("fail")
	}

	if ticks[20].Close.Cmp(&testTrades[2].Price) != 0 {
		t.Errorf("fail")
	}
	if ticks[20].High.Cmp(&testTrades[2].Price) != 0 {
		t.Errorf("fail")
	}

	if ticks[20].Low.Cmp(&testTrades[2].Price) != 0 {
		t.Errorf("fail")
	}

	if ticks[20].Volume.Cmp(big.NewFloat(0)) != 0 {
		t.Errorf("fail")
	}

	if ticks[20].LastOriginID != testTrades[2].OriginID {
		t.Errorf("fail")
	}

	fmt.Println("new ticks", len(ticks))
	fmt.Println("ticks[0]:")
	util.PrintTick(&ticks[0])
	fmt.Println("ticks[10]:")
	util.PrintTick(&ticks[10])
	fmt.Println("ticks[20]:")
	util.PrintTick(&ticks[20])
	fmt.Println("ticks[31]:")
	util.PrintTick(&ticks[31])
	fmt.Println("CurrentTick")
	util.PrintTick(&granularityObjectTick.CurrentTick)

	// add forth trade
	ticks = trade.Granulate(testTrades[4], &granularityObjectTick)
	if len(ticks) == 0 {
		t.Errorf("fail")
	}

	if granularityObjectTick.CurrentTick.Close.Cmp(&testTrades[4].Price) != 0 {
		t.Errorf("fail")
	}

	if granularityObjectTick.CurrentTick.High.Cmp(&testTrades[4].Price) != 0 {
		t.Errorf("high should be %s but was %s", testTrades[4].Price.String(), granularityObjectTick.CurrentTick.High.String())
	}

	if granularityObjectTick.CurrentTick.Open.Cmp(&testTrades[4].Price) != 0 {
		t.Errorf("fail")
	}

	if granularityObjectTick.CurrentTick.Low.Cmp(&testTrades[4].Price) != 0 {
		t.Errorf("fail")
	}

	if granularityObjectTick.CurrentTick.Volume.Cmp(&testTrades[4].Amount) != 0 {
		t.Errorf("fail")
	}

	if granularityObjectTick.CurrentTick.LastOriginID != testTrades[4].OriginID {
		t.Errorf("fail")
	}

	if ticks[20].Open.Cmp(&testTrades[3].Price) != 0 {
		t.Errorf("fail")
	}

	if ticks[20].Close.Cmp(&testTrades[3].Price) != 0 {
		t.Errorf("fail")
	}
	if ticks[20].High.Cmp(&testTrades[3].Price) != 0 {
		t.Errorf("fail")
	}

	if ticks[20].Low.Cmp(&testTrades[3].Price) != 0 {
		t.Errorf("fail")
	}

	if ticks[20].Volume.Cmp(big.NewFloat(0)) != 0 {
		t.Errorf("fail")
	}

	if ticks[20].LastOriginID != testTrades[3].OriginID {
		t.Errorf("fail")
	}

	fmt.Println("new ticks", len(ticks))
	fmt.Println("ticks[0]:")
	util.PrintTick(&ticks[0])
	fmt.Println("ticks[10]:")
	util.PrintTick(&ticks[10])
	fmt.Println("ticks[20]:")
	util.PrintTick(&ticks[20])
	fmt.Println("ticks[299]:")
	util.PrintTick(&ticks[299])
	fmt.Println("CurrentTick")
	util.PrintTick(&granularityObjectTick.CurrentTick)
}
