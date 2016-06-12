package database

import (
	"database/sql"
	"fmt"
	"strconv"
	"github.com/cluda/btcdata/util"

	"github.com/cluda/btcdata/trade"
	_ "github.com/lib/pq"
)

// CreateTradeTableIfNotExcists creates the trade table if it does not already exciste
func CreateTradeTableIfNotExcists(db *sql.DB) (string, error) {
	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS bitfinex_trade (
  id serial primary key,
  origin_id bigint NOT NULL,
  trade_time bigint NOT NULL,
  price numeric(10,3) NOT NULL,
  amount numeric(20,8) NOT NULL,
  trade_type varchar(5) NOT NULL
  )`)
	if err != nil {
		fmt.Println("EROR: failed when trying to create table")
		return "", err
	}
	return "OK", nil
}

// GetTrades returns trads afther 'afterOriginID'
// the first trade is the oldest and the last is the newest
func GetTrades(db *sql.DB, afterOriginID int64) ([]trade.Trade, error) {
	//fmt.Println("strconv.FormatInt(afterOriginID, 10):", )

	var trades []trade.Trade

	var (
		id        int64
		originID  int64
		tradeTime int64
		price     float64
		amount    float64
		typeTrade string
	)

	rows, err := db.Query("SELECT id, origin_id, trade_time, price, amount, trade_type from bitfinex_trade WHERE origin_id > " + strconv.FormatInt(afterOriginID, 10) + " order by origin_id")
	if err != nil {
		fmt.Println("ERROR: failed to get rows of trads")
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		err := rows.Scan(&id, &originID, &tradeTime, &price, &amount, &typeTrade)
		if err != nil {
			fmt.Println("ERROR: getTradesAfter failed on rows.Next()")
			return nil, err
		}
		trades = append(trades, trade.Trade{
			ID:        id,
			OriginID:  originID,
			TradeTime: tradeTime,
			Price:     price,
			Amount:    amount,
			Type:      typeTrade,
		})
	}
	err = rows.Err()
	if err != nil {
		fmt.Println("ERROR: getTradesAfter failed on rows.Err()")
		return nil, err
	}
	return trades, nil
}

// GetFirstTradeAfther returns the first trade that has a originID higher then this
func GetFirstTradeAfther(db *sql.DB, originID int64) (trade.Trade, error) {
	var thisTrade = trade.Trade{}
	err := db.QueryRow("SELECT id, origin_id, trade_time, price, amount, trade_type from bitfinex_trade WHERE origin_id > "+strconv.FormatInt(originID, 10)+" order by origin_id limit 1").Scan(&thisTrade.ID, &thisTrade.OriginID, &thisTrade.TradeTime, &thisTrade.Price, &thisTrade.Amount, &thisTrade.Type)
	if err != nil {
		fmt.Printf("could not get trade after %v\n", originID)
		return thisTrade, err
	}
	return thisTrade, nil
}

// GetOldestTrade will return the first trade in the trade table
func GetOldestTrade(db *sql.DB) (trade.Trade, error) {
	var thisTrade = trade.Trade{}
	err := db.QueryRow("SELECT id, origin_id, trade_time, price, amount, trade_type from bitfinex_trade order by origin_id limit 1").Scan(&thisTrade.ID, &thisTrade.OriginID, &thisTrade.TradeTime, &thisTrade.Price, &thisTrade.Amount, &thisTrade.Type)
	if err != nil {
		fmt.Println("could not get the first trade from the trade table")
		return thisTrade, err
	}
	return thisTrade, nil
}

// GetFirstNewest will return the first trade in the trade table
func GetNewestTrade(db *sql.DB) (trade.Trade, error) {
	var thisTrade = trade.Trade{}
	err := db.QueryRow("SELECT id, origin_id, trade_time, price, amount, trade_type from bitfinex_trade order by origin_id desc limit 1").Scan(&thisTrade.ID, &thisTrade.OriginID, &thisTrade.TradeTime, &thisTrade.Price, &thisTrade.Amount, &thisTrade.Type)
	if err != nil {
		fmt.Println("could not get the first trade from the trade table")
		return thisTrade, err
	}
	return thisTrade, nil
}

// InsertTrades will insert the trades and return a string
// the trades table should have the newest trade first and the oldest trade last
func InsertTrades(db *sql.DB, trades []trade.Trade) (string, error) {
	if len(trades) > 0 {
		sqlStr := "INSERT INTO bitfinex_trade (origin_id, trade_time, price, amount, trade_type) VALUES "

		for i := len(trades) - 1; i >= 0; i-- {
			sqlStr += "(" + strconv.FormatInt(trades[i].OriginID, 10) + ", " + strconv.FormatInt(trades[i].TradeTime, 10) + ", " + util.PriceToString(trades[i].Price) + ", " + util.AmountToString(trades[i].Amount) + ", '" + trades[i].Type + "'),"
		}

		//trim the last ,
		sqlStr = sqlStr[0:len(sqlStr)-1] + ";"

		//write to database
		_, err := db.Exec(sqlStr)
		if err != nil {
			fmt.Println("could not write th trades to the database")
			return "", err
		}
		return "OK", nil
	}

	return "OK", nil
}
