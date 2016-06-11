package database

import (
	"database/sql"
	"fmt"
	"math/big"
	"strconv"

	"github.com/cluda/btcdata/trade"
	_ "github.com/lib/pq"
)

func CreateTickTablesForIntervalls(db *sql.DB, intervalls []int) (string, error) {
	for _, value := range intervalls {
		_, err := db.Exec("CREATE TABLE IF NOT EXISTS bitfinex_tick_" + strconv.Itoa(value) + ` (
  id serial primary key,
  open numeric(10,3) NOT NULL,
  close numeric(10,3) NOT NULL,
  high numeric(10,3) NOT NULL,
  low numeric(10,3) NOT NULL,
  volume numeric(20,8) NOT NULL,
  last_origin_id bigint NOT NULL,
  tick_end_time bigint NOT NULL
  )`)
		if err != nil {
			fmt.Println("could not create table for granularitie ", value)
			return "", err
		}
	}
	return "OK", nil
}

func InsertTicks(db *sql.DB, tableName string, ticks []trade.Tick) (string, error) {
	if len(ticks) == 0 {
		return "NO TICKS", nil
	}

	sqlStr := "INSERT INTO " + tableName + " (open, close, high, low, volume, last_origin_id, tick_end_time) VALUES "

	for i := len(ticks) - 1; i >= 0; i-- {
		tick := (ticks)[i]
		sqlStr += "(" + tick.Open.String() + ", " + tick.Close.String() + ", " + tick.High.String() + ", " + tick.Low.String() + ", " + tick.Volume.String() + ", " + strconv.FormatInt(tick.LastOriginID, 10) + ", " + strconv.FormatInt(tick.TickEndTime, 10) + "),"
	}
	//trim the last ,
	sqlStr = sqlStr[0:len(sqlStr)-1] + ";"

	//write to database
	_, err := db.Exec(sqlStr)
	if err != nil {
		fmt.Println("addTrade: could not write the ticks to the database")
		return "", err
	}
	return "OK", nil
}

// GetLastTickIfAnyForIntervalls returns a map with last tick in the database for the intervalls
// if a intervall has no ticks 
func GetLastTickIfAnyForIntervalls(db *sql.DB, intervalls []int) map[int]trade.Tick {
	tickMap := make(map[int]trade.Tick)

	var (
		openStr      string
		closeStr     string
		highStr      string
		lowStr       string
		volumeStr    string
		lastOriginID int64
		tickEndTime  int64
	)

	for _, interval := range intervalls {
		err := db.QueryRow("SELECT open, close, high, low, volume, last_origin_id, tick_end_time from bitfinex_tick_"+strconv.Itoa(interval)+" order by last_origin_id desc limit 1").Scan(&openStr, &closeStr, &highStr, &lowStr, &volumeStr, &lastOriginID, &tickEndTime)
		if err != nil {
			fmt.Printf("no last tick for %v interval. Continues to next. \n", interval)
		} else {
			open, _ := new(big.Float).SetString(openStr)
			close, _ := new(big.Float).SetString(closeStr)
			high, _ := new(big.Float).SetString(highStr)
			low, _ := new(big.Float).SetString(lowStr)
			volume, _ := new(big.Float).SetString(volumeStr)

			tickMap[interval] = trade.Tick{
				Open:         *open,
				Close:        *close,
				High:         *high,
				Low:          *low,
				Volume:       *volume,
				LastOriginID: lastOriginID,
				TickEndTime:  tickEndTime,
			}
		}
	}
	return tickMap
}
