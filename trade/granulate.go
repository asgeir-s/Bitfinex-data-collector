package trade

import (
	"math/big"
)

// Granularity is a helper object for building Tick for a spesific granularity
// .granularity is in secounds
type Granularity struct {
	interval    int64
	tableName   string
	currentTick *Tick
}

// Granulate trade into granularity and returns new Tick if therie is a new Tick
func Granulate(trade Trade, granularity *Granularity) []*Tick {
	newTicks := []*Tick{}
	if trade.OriginID > granularity.currentTick.LastOriginID {
		for trade.TradeTime > granularity.currentTick.TickEndTime {
			// add tick to return table
			newTicks = append(newTicks, granularity.currentTick)
			granularity.currentTick = &Tick{
				Open:         granularity.currentTick.Close,
				Close:        granularity.currentTick.Close,
				High:         granularity.currentTick.Close,
				Low:          granularity.currentTick.Close,
				Volume:       *new(big.Float),
				LastOriginID: granularity.currentTick.LastOriginID,
				TickEndTime:  granularity.currentTick.TickEndTime + granularity.interval,
			}
		}
		addTradeToTick(trade, granularity.currentTick)
	}
	return newTicks
}

// InitializeGranularityFromTick should be used when continuing building a already started Tick-table
func InitializeGranularityFromTick(lastTick Tick, tableName string, interval int64) Granularity {
	return Granularity{
		interval:  interval,
		tableName: tableName,
		currentTick: &Tick{
			Open:         lastTick.Close,
			Close:        lastTick.Close,
			High:         lastTick.Close,
			Low:          lastTick.Close,
			Volume:       *new(big.Float),
			LastOriginID: lastTick.LastOriginID,
			TickEndTime:  lastTick.TickEndTime + interval,
		},
	}
}

// InitializeGranularityFromTrade should only be used when creating the first tick
// if theire already is Ticks in the table use 'InitializeGranularityFromTick'
func InitializeGranularityFromTrade(trade Trade, tableName string, interval int64) Granularity {
	return Granularity{
		interval:  interval,
		tableName: tableName,
		currentTick: &Tick{
			Open:         trade.Price,
			Close:        trade.Price,
			High:         trade.Price,
			Low:          trade.Price,
			Volume:       *new(big.Float),
			LastOriginID: trade.OriginID,
			TickEndTime:  trade.TradeTime + interval,
		},
	}
}

func addTradeToTick(trade Trade, tick *Tick) {
	tick.LastOriginID = trade.OriginID
	if trade.Price.Cmp(&tick.High) > 0 {
		tick.High = trade.Price
	} else if trade.Price.Cmp(&tick.Low) < 0 {
		tick.Low = trade.Price
	}
	tick.Volume.Add(&tick.Volume, &trade.Amount)
	tick.Close = trade.Price
}
