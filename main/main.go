package main

import (
	"database/sql"
	"fmt"
	"log"

	"strconv"

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

	_, err = db.Exec("SET TIME ZONE 'UTC';")
	if err != nil {
		fmt.Println("could not set timezone on database")
		log.Fatal(err)
	}

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

		for i := len(trades) - 1; i >= 0; i-- {
			trade := trades[i]
			sqlStr += "(" + strconv.FormatInt(trade.TradeId, 10) + ", " + strconv.FormatInt(trade.Timestamp, 10) + ", " + trade.Price.String() + ", " + trade.Amount.String() + ", '" + trade.Type + "'),"
		}

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
