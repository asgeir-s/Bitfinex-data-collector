package trade

import (
	"strconv"
	"math/big"
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
				Volume:       *new(big.Float),
				LastOriginID: granularity.CurrentTick.LastOriginID,
				TickEndTime:  granularity.CurrentTick.TickEndTime + granularity.Interval,
			}
		}
		if granularity.CurrentTick.Volume.Cmp(big.NewFloat(0)) == 0 {
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
			addTradeToTick(trade, &granularity.CurrentTick)
		}
	}
	return newTicks
}

// InitializeGranularityFromTick should be used when continuing building a already started Tick-table
func InitializeGranularityFromTick(lastTick Tick, tableName string, interval int64) Granularity {
	return Granularity{
		Interval:  interval,
		TableName: tableName,
		CurrentTick: Tick{
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

// InitializeGranularities re-initializes alreadt started granularities and initializes new granularities
func InitializeGranularities(intervalls []int, lastTicks map[int]Tick, oldestProsesedTrade Trade) map[int]Granularity {
	granularityMap := make(map[int]Granularity)

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