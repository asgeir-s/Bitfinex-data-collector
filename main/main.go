package main

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/caarlos0/env"
	"github.com/cluda/bitfinex-api-go"
	"github.com/cluda/btcdata/database"
	"github.com/cluda/btcdata/trade"
	"github.com/cluda/btcdata/util"
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
	granularitiInterval := []int{1800, 3600} //1800, 3600, 7200, 14400, 21600, 28800, 43200, 86400

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

	// create table if not exists
	_, err = database.CreateTradeTableIfNotExcists(db)
	if err != nil {
		fmt.Println("could not create table")
		log.Fatal(err)
	}
	_, err = database.CreateTickTablesForIntervalls(db, granularitiInterval)
	if err != nil {
		fmt.Println("could not create granularitie-tables")
		log.Fatal(err)
	}

	// get new trades
	var newestTradeTime int64
	newestTrade, err := database.GetNewestTrade(db)
	if err != nil {
		fmt.Println("could not get newest trade from trade table. newestTradeTime is 0 ")
	} else {
		newestTradeTime = newestTrade.TradeTime
	}
	println("time of the newest trade in the trade table is", newestTradeTime)

	bfxClient := bitfinex.NewClient()
	bfxTrades, err := bfxClient.Trades.All("btcusd", newestTradeTime+1, 0)
	if err != nil {
		fmt.Println("could not get trades from the bitfinex rest api. Will retry.")
		bfxTrades, err = bfxClient.Trades.All("btcusd", newestTradeTime+1, 0)
		if err != nil {
			fmt.Println("could not get trades from the bitfinex rest api. Failed on retry.")
			log.Fatal(err)
		}
	}

	newTrades := util.BitfinexTradetoTrade(bfxTrades)
	fmt.Println("got", len(newTrades), "new trades from bitfinex")

	_, err = database.InsertTrades(db, newTrades)
	if err != nil {
		fmt.Println("could not get insert trades to the trade table")
		log.Fatal(err)
	}

	lastTicks := database.GetLastTickIfAnyForIntervalls(db, granularitiInterval)
	oldestTickOriginID := getOldestOriginID(granularitiInterval, lastTicks)
	fmt.Println("oldestTickOriginId", oldestTickOriginID)

	tradesThatNeedGranulating, err := database.GetTrades(db, oldestTickOriginID)
	if err != nil {
		fmt.Println("could not create granularitie-tables")
		log.Fatal(err)
	}
	granularities := trade.InitializeGranularities(granularitiInterval, lastTicks, tradesThatNeedGranulating[0])

	granulateTrades := func(thisTrades []trade.Trade) {
		for _, thisTrade := range thisTrades {
			for _, interval := range granularitiInterval {
				ticks := trade.Granulate(thisTrade, granularities[interval])
				database.InsertTicks(db, granularities[interval].TableName, ticks)
			}
		}
	}
	granulateTrades(tradesThatNeedGranulating)

	// Create new connection
	err = bfxClient.WebSocket.Connect()
	if err != nil {
		log.Fatal("Error connecting to web socket")
	}
	defer bfxClient.WebSocket.Close()

	tradesChan := make(chan []float64)

	bfxClient.WebSocket.AddSubscribe(bitfinex.CHAN_TRADE, bitfinex.BTCUSD, tradesChan)
	go bfxClient.WebSocket.Subscribe()

	// after api client successfully connect to remote web socket
	// channel will reveive current payload as separate messages.
	// each channel will receive order book updates: [price, count, ±amount]
	for {
		select {
		case tradeMsg := <-tradesChan:
			thisTrade := []trade.Trade{util.BitfinexWSTradeArrayToTrade(tradeMsg)}
			// add to tick table
			_, err = database.InsertTrades(db, thisTrade)
			if err != nil {
				log.Fatalln("could not add this trade to the trade table. Error: ", err.Error())
			}
			// granulate
			granulateTrades(thisTrade)
			log.Printf("%#v\n", thisTrade[0])
		}
	}

}

func getOldestOriginID(intervalls []int, ticks map[int]trade.Tick) int64 {
	var oldest int64
	for _, value := range intervalls {
		this, exists := ticks[value]
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
