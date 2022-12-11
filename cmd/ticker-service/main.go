package main

import (
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/endk17/go-crypto-ticker/cmd/ticker-service/config"
	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/k0kubun/pp/v3"
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
	log.Println("Listing for requests at http://localhost:8000/ping")
	log.Fatal(http.ListenAndServe(":8000", nil))

	cfg, err := config.Read()
	if err != nil {
		log.Fatal().Err(err).Send()
	}
	pp.Println(cfg)

	// Create new influxdb client with default option for server url authenticate by token
	influxdb := influxdb2.NewClient(cfg.InfluxDB.URL, cfg.InfluxDB.Token)
	defer influxdb.Close()
}
