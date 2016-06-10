package main

import (
	"database/sql"
	"fmt"
	"log"
	"math/big"
	"strconv"

	"github.com/cluda/btcdata/trade"

	"github.com/caarlos0/env"
	"github.com/cluda/bitfinex-api-go"
	_ "github.com/lib/pq"
)

type databaseConfig struct {
	Host    string `env:"DB_HOST" envDefault:"localhost"`
	User    string `env:"DB_USER" envDefault:"testuser"`
	Pasword string `env:"DB_PASSWORD" envDefault:"Password123"`
	Name    string `env:"DB_NAME" envDefault:"timeseries"`
	Port    string `env:"DB_PORT" envDefault:"5432"`
}

func main() {

	granularitiInterval := []int{7200} //1800, 3600, 7200, 14400, 21600, 28800, 43200, 86400

	dbConfig := databaseConfig{}
	env.Parse(&dbConfig)

	db, err := sql.Open("postgres",
		"host="+dbConfig.Host+
			" user="+dbConfig.User+
			" password="+dbConfig.Pasword+
			" dbname="+dbConfig.Name+
			" port="+dbConfig.Port+
			" sslmode=disable")
	if err != nil {
		fmt.Println("could not connect to the databse")
		log.Fatal(err)
	}
	defer db.Close()

	// check trade table existc if not create
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS bitfinex_trade (
  id serial primary key,
  origin_id bigint NOT NULL,
  trade_time bigint NOT NULL,
  price numeric(10,3) NOT NULL,
  amount numeric(20,8) NOT NULL,
  trade_type varchar(5) NOT NULL
  )`)
	if err != nil {
		fmt.Println("could not create table")
		log.Fatal(err)
	}

	_, err = createGranularitiesTables(db, granularitiInterval)
	if err != nil {
		fmt.Println("could not create granularitie-tables")
		log.Fatal(err)
	}

	lastTicks := getLastTickIfAny(db, granularitiInterval)
	oldestTickOriginTradeID := getOldestTradeOriginIDInTickTables(granularitiInterval, lastTicks)

	fmt.Println("oldestTickOriginTradeID", oldestTickOriginTradeID)

	var oldestProsesedTrade = trade.Trade{}
	var priceStr string
	var amountStr string
	err = db.QueryRow("SELECT id, origin_id, trade_time, price, amount, trade_type from bitfinex_trade WHERE origin_id > "+strconv.FormatInt(oldestTickOriginTradeID, 10)+" order by origin_id limit 1").Scan(&oldestProsesedTrade.ID, &oldestProsesedTrade.OriginID, &oldestProsesedTrade.TradeTime, &priceStr, &amountStr, &oldestProsesedTrade.Type)
	if err != nil {
		fmt.Println("could not get time of last trade in database. WIll use zero value")
		fmt.Println(err)
	} else {
		price, _ := new(big.Float).SetString(priceStr)
		amount, _ := new(big.Float).SetString(amountStr)
		oldestProsesedTrade.Price = *price
		oldestProsesedTrade.Amount = *amount
	}

	granularities := getGranularityes(granularitiInterval, lastTicks, oldestProsesedTrade)

	for _, value := range granularities {
		printTick(value.CurrentTick)
	}

	// granulate tick from db
	tradsThatNeedGranulating := getTradesAfter(db, oldestTickOriginTradeID)

	// get all trades and granulate them
	for _, thisTrade := range tradsThatNeedGranulating {
		addTrade(db, thisTrade, &granularities)
	}

	// get last tick for all tick tables

	var newestTradeTime int64

	err = db.QueryRow("SELECT trade_time from bitfinex_trade order by trade_time desc limit 1").Scan(&newestTradeTime)
	if err != nil {
		fmt.Println("could not get time of last trade in database. WIll use zero value")
	}

	// get last trade in DB
	client := bitfinex.NewClient()

	// get last trades ather "last trade in DB" from Bitfinex REST
	trades, err := client.Trades.All("btcusd", newestTradeTime+1, 0)
	if err != nil {
		fmt.Println("could not get trades from the bitfinex rest api. Will retry.")
		trades, err = client.Trades.All("btcusd", newestTradeTime+1, 0)
		if err != nil {
			fmt.Println("could not get trades from the bitfinex rest api. Failed on retry.")
			log.Fatal(err)
		}
	}

	fmt.Println("got", len(trades), "new trades from bitfinex")

	if len(trades) > 0 {
		sqlStr := "INSERT INTO bitfinex_trade (origin_id, trade_time, price, amount, trade_type) VALUES "

		for i := 0; i <= len(trades)-1; i++ {
			thisTrade := trade.Trade{
				OriginID:  trades[i].TradeId,
				TradeTime: trades[i].Timestamp,
				Price:     trades[i].Price,
				Amount:    trades[i].Amount,
				Type:      trades[i].Type,
			}
			printTrade(&thisTrade)

			sqlStr += "(" + strconv.FormatInt(thisTrade.OriginID, 10) + ", " + strconv.FormatInt(thisTrade.TradeTime, 10) + ", " + thisTrade.Price.String() + ", " + thisTrade.Amount.String() + ", '" + thisTrade.Type + "'),"
			addTrade(db, thisTrade, &granularities)
		}

    println("sqlStr:", sqlStr)
		//trim the last ,
		sqlStr = sqlStr[0:len(sqlStr)-1] + ";"

		//write to database
		res, err := db.Exec(sqlStr)
		if err != nil {
			fmt.Println("could not write th trades to the database")
			log.Fatal(err)
		}

		fmt.Println("res:", res)
	}

	//trade.Granulate()
	// granulate on write to DB

	// get new trades from Bitfinex rest again (in case it took long to get it into DB)

	// rubscribe to websocket and granulate on save
}

func createGranularitiesTables(db *sql.DB, granularities []int) (string, error) {
	for _, value := range granularities {
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

func getOldestTradeOriginIDInTickTables(granularities []int, lastTicks map[int]trade.Tick) int64 {
	var oldest int64
	for _, value := range granularities {
		this, exists := lastTicks[value]
		if !exists {
			return 0
		} else if oldest == 0 {
			oldest = this.LastOriginID
		} else if this.LastOriginID < oldest {
			oldest = this.LastOriginID
		}
	}
	return oldest
}

func getLastTickIfAny(db *sql.DB, granularities []int) map[int]trade.Tick {

	tickMap := make(map[int]trade.Tick)

	var openStr string
	var closeStr string
	var highStr string
	var lowStr string
	var volumeStr string
	var lastOriginID int64
	var tickEndTime int64

	for _, value := range granularities {
		err := db.QueryRow("SELECT open, close, high, low, volume, last_origin_id, tick_end_time from bitfinex_tick_"+strconv.Itoa(value)+" order by last_origin_id desc limit 1").Scan(&openStr, &closeStr, &highStr, &lowStr, &volumeStr, &lastOriginID, &tickEndTime)
		if err != nil {
			fmt.Println("could not get last tick from database. Continues to next. (this is not a problem)")
		} else {
			open, _ := new(big.Float).SetString(openStr)
			close, _ := new(big.Float).SetString(closeStr)
			high, _ := new(big.Float).SetString(highStr)
			low, _ := new(big.Float).SetString(lowStr)
			volume, _ := new(big.Float).SetString(volumeStr)

			tickMap[value] = trade.Tick{
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

func getGranularityes(granularitiIntervals []int, lastTicks map[int]trade.Tick, oldestProsesedTrade trade.Trade) map[int]trade.Granularity {
	granularityMap := make(map[int]trade.Granularity)

	for _, value := range granularitiIntervals {
		tick, exists := lastTicks[value]
		if !exists {
			granularityMap[value] = trade.InitializeGranularityFromTrade(oldestProsesedTrade, "bitfinex_tick_"+strconv.Itoa(value), int64(value))
		} else {
			granularityMap[value] = trade.InitializeGranularityFromTick(tick, "bitfinex_tick_"+strconv.Itoa(value), int64(value))
		}
	}
	return granularityMap
}

func getTradesAfter(db *sql.DB, afterOriginID int64) []trade.Trade {
	//fmt.Println("strconv.FormatInt(afterOriginID, 10):", )

	var trades []trade.Trade

	var (
		id        int64
		originID  int64
		tradeTime int64
		priceStr  string
		amountStr string
		typeTrade string
	)

	rows, err := db.Query("SELECT id, origin_id, trade_time, price, amount, trade_type from bitfinex_trade WHERE origin_id > " + strconv.FormatInt(afterOriginID, 10) + " order by origin_id")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	for rows.Next() {
		err := rows.Scan(&id, &originID, &tradeTime, &priceStr, &amountStr, &typeTrade)
		if err != nil {
			fmt.Println("getTradesAfter failed on rows.Next()")
			log.Fatal(err)
		} else {
			price, _ := new(big.Float).SetString(priceStr)
			amount, _ := new(big.Float).SetString(amountStr)
			trades = append(trades, trade.Trade{
				ID:        id,
				OriginID:  originID,
				TradeTime: tradeTime,
				Price:     *price,
				Amount:    *amount,
				Type:      typeTrade,
			})
		}
	}
	err = rows.Err()
	if err != nil {
		fmt.Println("getTradesAfter failed on rows.Err()")
		log.Fatal(err)
	}
	return trades
}

func addTrade(db *sql.DB, thisTrade trade.Trade, granularities *map[int]trade.Granularity) {
	//fmt.Println("thisTrade.OriginID:", thisTrade.OriginID)
	for _, granularity := range *granularities {
		ticks := trade.Granulate(thisTrade, &granularity)

		// write ticks to database
		if len(ticks) > 0 {
			sqlStr := "INSERT INTO " + granularity.TableName + " (open, close, high, low, volume, last_origin_id, tick_end_time) VALUES "

			for i := len(ticks) - 1; i >= 0; i-- {
				tick := ticks[i]
				//printTick(tick)
				sqlStr += "(" + tick.Open.String() + ", " + tick.Close.String() + ", " + tick.High.String() + ", " + tick.Low.String() + ", " + tick.Volume.String() + ", " + strconv.FormatInt(tick.LastOriginID, 10) + ", " + strconv.FormatInt(tick.TickEndTime, 10) + "),"
			}
			//trim the last ,
			sqlStr = sqlStr[0:len(sqlStr)-1] + ";"

			//write to database
			_, err := db.Exec(sqlStr)
			if err != nil {
				fmt.Println("addTrade: could not write the ticks to the database")
				log.Fatal(err)
			}
		}
	}
}

func printTick(tick *trade.Tick) {
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

func printTrade(trade *trade.Trade) {
	fmt.Println("{")
	fmt.Println("   ID: ", trade.ID)
	fmt.Println("   OriginID: ", trade.OriginID)
	fmt.Println("   Price: ", trade.Price.String())
	fmt.Println("   Amount: ", trade.Amount.String())
	fmt.Println("   TradeTime: ", trade.TradeTime)
	fmt.Println("   Type: ", trade.Type)
	fmt.Println("},")
}
