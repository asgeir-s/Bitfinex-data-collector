package trade

import (
	"strconv"
)

// Granularity is a helper object for building Tick for a spesific granularity
// .granularity is in secounds
type Granularity struct {
	Interval    int64
	TableName   string
	CurrentTick Tick
}

// Granulate trade into granularity and returns new Tick if therie is a new Tick
func Granulate(trade Trade, granularity *Granularity) []Tick {
	newTicks := []Tick{}

	if trade.OriginID > granularity.CurrentTick.LastOriginID {
		for trade.TradeTime > granularity.CurrentTick.TickEndTime {
			// add tick to return table
			newTicks = append(newTicks, granularity.CurrentTick)
			granularity.CurrentTick = Tick{
				Open:         granularity.CurrentTick.Close,
				Close:        granularity.CurrentTick.Close,
				High:         granularity.CurrentTick.Close,
				Low:          granularity.CurrentTick.Close,
				Volume:       0,
				LastOriginID: granularity.CurrentTick.LastOriginID,
				TickEndTime:  granularity.CurrentTick.TickEndTime + granularity.Interval,
			}
		}
		if granularity.CurrentTick.Volume == 0 {
			granularity.CurrentTick = Tick{
				Open:         trade.Price,
				Close:        trade.Price,
				High:         trade.Price,
				Low:          trade.Price,
				Volume:       trade.Amount,
				LastOriginID: trade.OriginID,
				TickEndTime:  granularity.CurrentTick.TickEndTime,
			}
		} else {
			addTradeToGranularity(trade, &granularity.CurrentTick)
		}
	}
	return newTicks
}

// InitializeGranularityFromTick should be used when continuing building a already started Tick-table
func InitializeGranularityFromTick(lastTick Tick, tableName string, interval int64) *Granularity {
	return &Granularity{
		Interval:  interval,
		TableName: tableName,
		CurrentTick: Tick{
			Open:         lastTick.Close,
			Close:        lastTick.Close,
			High:         lastTick.Close,
			Low:          lastTick.Close,
			Volume:       0,
			LastOriginID: lastTick.LastOriginID,
			TickEndTime:  lastTick.TickEndTime + interval,
		},
	}
}

// InitializeGranularityFromTrade should only be used when creating the first tick
// if theire already is Ticks in the table use 'InitializeGranularityFromTick'
func InitializeGranularityFromTrade(trade Trade, tableName string, interval int64) *Granularity {
	return &Granularity{
		Interval:  interval,
		TableName: tableName,
		CurrentTick: Tick{
			Open:         trade.Price,
			Close:        trade.Price,
			High:         trade.Price,
			Low:          trade.Price,
			Volume:       trade.Amount,
			LastOriginID: trade.OriginID,
			TickEndTime:  trade.TradeTime + interval,
		},
	}
}

func addTradeToGranularity(trade Trade, tick *Tick) {
	tick.LastOriginID = trade.OriginID
	if trade.Price > tick.High {
		tick.High = trade.Price
	} else if trade.Price < tick.Low {
		tick.Low = trade.Price
	}
	tick.Volume += trade.Amount
	tick.Close = trade.Price
}

// InitializeGranularities re-initializes alreadt started granularities and initializes new granularities
func InitializeGranularities(intervalls []int, lastTicks map[int]Tick, oldestProsesedTrade Trade) map[int]*Granularity {
	granularityMap := make(map[int]*Granularity)

	for _, value := range intervalls {
		tick, exists := lastTicks[value]
		if !exists {
			granularityMap[value] = InitializeGranularityFromTrade(oldestProsesedTrade, "bitfinex_tick_"+strconv.Itoa(value), int64(value))
		} else {
			granularityMap[value] = InitializeGranularityFromTick(tick, "bitfinex_tick_"+strconv.Itoa(value), int64(value))
		}
	}
	return granularityMap
}
