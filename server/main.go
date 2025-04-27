package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/bongofriend/torrent-ingest/api"
	"github.com/bongofriend/torrent-ingest/config"
	"github.com/bongofriend/torrent-ingest/torrent"
)

var configFilePath string

func main() {
	parseFlags()
	appConfig, err := config.LoadConfig(configFilePath)
	if err != nil {
		panic(err)
	}

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGABRT, syscall.SIGINT)

	torrentService, err := torrent.NewTorrentService(appConfig.Torrent)
	if err != nil {
		panic(err)
	}
	if err := torrentService.Init(); err != nil {
		panic(err)
	}
	appContext, cancel := context.WithCancel(context.Background())
	wg := &sync.WaitGroup{}

	wg.Add(1)
	go api.StartServer(appContext, wg, appConfig, torrentService)

	sig := <-signalChan
	log.Printf("Signal received: %s", sig)
	cancel()
	wg.Wait()
}

func parseFlags() {
	flag.StringVar(&configFilePath, "configFilePath", "./config.yml", "Path to config file")
	flag.Parse()
}
