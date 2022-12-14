package main

import (
	"context"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/endk17/go-crypto-ticker/cmd/ticker-service/config"
	"github.com/gorilla/websocket"
	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/k0kubun/pp/v3"
	"github.com/preichenberger/go-coinbasepro/v2"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func init() {
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339})
}

func main() {
	// Initial Set up: web server
	pingHandler := func(w http.ResponseWriter, req *http.Request) {
		io.WriteString(w, "pong!\n")
	}
	http.HandleFunc("/ping", pingHandler)
	pp.Println("Listing for requests at http://localhost:8000/ping")

	cfg, err := config.Read()
	if err != nil {
		log.Fatal().Err(err).Send()
	}
	pp.Println(cfg)

	// Create new influxdb client auth by token
	influxdb := influxdb2.NewClient(cfg.InfluxDB.URL, cfg.InfluxDB.Token)
	defer influxdb.Close()

	// User blocking write to the desired bucket
	writeAPI := influxdb.WriteAPIBlocking(cfg.InfluxDB.Org, cfg.InfluxDB.Bucket)

	// Create a websocket to coinbase
	var wsDialer websocket.Dialer
	ws, _, err := wsDialer.Dial("wss://ws-feed.pro.coinbase.com", nil)
	if err != nil {
		log.Fatal().Err(err).Send()
	}

	// sub msg
	subscribe := coinbasepro.Message{
		Type:       "subscribe",
		ProductIds: []string{"BTC-USD", "ETH-USD"},
		Channels: []coinbasepro.MessageChannel{
			{
				Name: "ticker",
			},
		},
	}

	// subscribe
	if err := ws.WriteJSON(subscribe); err != nil {
		log.Fatal().Err(err).Send()
	}

	// read tick data from ws
	messages := make(chan coinbasepro.Message)
	go func() {
		defer close(messages)
		for {
			// Read tick data from websocket
			message := coinbasepro.Message{}
			if err := ws.ReadJSON(&message); err != nil {
				log.Error().Err(err).Send()
				break
			}
			log.Info().Str("product", message.ProductID).Str("price", message.Price).Send()

			messages <- message
		}
	}()

	// Write tick data to db
	for message := range messages {
		// Convert price to float
		price, err := strconv.ParseFloat(message.Price, 64)
		if err != nil {
			log.Error().Err(err).Send()
			continue
		}

		// create influx point using params constructor
		p := influxdb2.NewPoint("tick",
			map[string]string{"product": message.ProductID},
			map[string]interface{}{"price": price},
			time.Now())

		// write point
		if err := writeAPI.WritePoint(context.Background(), p); err != nil {
			log.Error().Err(err).Send()
		}
	}
}
