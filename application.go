package main

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sns"
	"github.com/caarlos0/env"
	"github.com/cluda/bitfinex-api-go"
	"github.com/cluda/btcdata/awsutil"
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
	dbConfig := databaseConfig{}
	env.Parse(&dbConfig)

	granularitiInterval := []int{1800, 3600, 7200, 14400, 21600, 28800, 43200, 86400} //1800, 3600, 7200, 14400, 21600, 28800, 43200, 86400
	snsTopicArn := map[int]string{
		1800:  "arn:aws:sns:us-east-1:525932482084:bitfinex_tick_1800",
		3600:  "arn:aws:sns:us-east-1:525932482084:bitfinex_tick_3600",
		7200:  "arn:aws:sns:us-east-1:525932482084:bitfinex_tick_7200",
		14400: "arn:aws:sns:us-east-1:525932482084:bitfinex_tick_14400",
		21600: "arn:aws:sns:us-east-1:525932482084:bitfinex_tick_21600",
		28800: "arn:aws:sns:us-east-1:525932482084:bitfinex_tick_28800",
		43200: "arn:aws:sns:us-east-1:525932482084:bitfinex_tick_43200",
		86400: "arn:aws:sns:us-east-1:525932482084:bitfinex_tick_86400",
	}

	bfxClient := bitfinex.NewClient()
	svc := sns.New(session.New(), aws.NewConfig().WithRegion("us-east-1"))

	db, err := sql.Open("postgres",
		"host="+dbConfig.Host+
			" user="+dbConfig.User+
			" password="+dbConfig.Pasword+
			" dbname="+dbConfig.Name+
			" port="+dbConfig.Port+
			" sslmode=disable")
	if err != nil {
		log.Fatal("could not connect to the databse. Error: " + err.Error())
	}
	defer db.Close()

	// create table if not exists
	_, err = database.CreateTradeTableIfNotExcists(db)
	if err != nil {
		log.Fatal("could not create table. Error:" + err.Error())
	}
	_, err = database.CreateTickTablesForIntervalls(db, granularitiInterval)
	if err != nil {
		log.Fatal("could not create granularitie-tables. Error: " + err.Error())
	}

	_, err = getNewTradesAndInsertToTable(db, bfxClient)
	if err != nil {
		log.Fatal("could not get new trades and insert to table. Error: " + err.Error())
	}

	newestTicks := database.GetNewestTickIfAnyForIntervalls(db, granularitiInterval)
	oldestOriginIDFromTick := getOldestOriginID(granularitiInterval, newestTicks)
	log.Println("oldestOriginIDFromTick", oldestOriginIDFromTick)

	tradesThatNeedGranulating, err := database.GetTrades(db, oldestOriginIDFromTick)
	if err != nil {
		log.Fatal("could not create granularitie-tables. Error: " + err.Error())
	}
	granularities := trade.InitializeGranularities(granularitiInterval, newestTicks, tradesThatNeedGranulating[0])

	granulateTrades := func(thisTrades []trade.Trade, live bool) {
		for _, thisTrade := range thisTrades {
			for _, interval := range granularitiInterval {
				ticks := trade.Granulate(thisTrade, granularities[interval])
				database.InsertTicks(db, granularities[interval].TableName, ticks)
				if live && len(ticks) > 0 {
					awsutil.SnsPublish(svc, ticks[0], snsTopicArn[interval])
					if err != nil {
						log.Println("failed to publish to SNS. Error: " + err.Error())
					}
				}
			}
		}
	}
	granulateTrades(tradesThatNeedGranulating, false)

	// Create new connection
	err = bfxClient.WebSocket.Connect()
	if err != nil {
		log.Fatal("Error connecting to web socket. Error: " + err.Error())
	}
	defer bfxClient.WebSocket.Close()

	tradesChan := make(chan []float64)
	bfxClient.WebSocket.AddSubscribe(bitfinex.CHAN_TRADE, bitfinex.BTCUSD, tradesChan)
	go bfxClient.WebSocket.Subscribe()

	newestOriginID := tradesThatNeedGranulating[len(tradesThatNeedGranulating)-1].OriginID
	for {
		select {
		case tradeMsg := <-tradesChan:
			thisTrade := []trade.Trade{util.BitfinexWSTradeArrayToTrade(tradeMsg)}
			if thisTrade[0].OriginID > newestOriginID {
				// add to tick table
				_, err = database.InsertTrades(db, thisTrade)
				if err != nil {
					log.Fatal("could not add this trade to the trade table. Error: ", err.Error())
				}
				// granulate
				granulateTrades(thisTrade, true)
				fmt.Printf(".")
				newestOriginID = thisTrade[0].OriginID
			} else {
				log.Printf("ignores already prosessed. tradeID: %v, lastProsessedTradeID: %v", thisTrade[0].OriginID, newestOriginID)
			}
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

// returns the number of new trades
func getNewTradesAndInsertToTable(db *sql.DB, bfxClient *bitfinex.Client) (int, error) {
	// get new trades
	var newestTradeTime int64
	newestTrade, err := database.GetNewestTrade(db)
	if err != nil {
		log.Println("could not get newest trade from trade table. newestTradeTime is 0 ")
	} else {
		newestTradeTime = newestTrade.TradeTime
	}
	log.Println("time of the newest trade in the trade table is", newestTradeTime)

	bfxTrades, err := bfxClient.Trades.All("btcusd", newestTradeTime+1, 0)
	if err != nil {
		log.Println("could not get trades from the bitfinex rest api. Will retry.")
		bfxTrades, err = bfxClient.Trades.All("btcusd", newestTradeTime+1, 0)
		if err != nil {
			log.Println("could not get trades from the bitfinex rest api. Failed on retry.")
			return 0, err
		}
	}

	newTrades := util.BitfinexTradetoTrade(bfxTrades)

	_, err = database.InsertTrades(db, newTrades)
	if err != nil {
		log.Println("could not get insert trades to the trade table")
		return 0, err
	}
	log.Printf("new trades form Bitfinex(REST): %v\n", len(newTrades))
	return len(newTrades), nil
}
