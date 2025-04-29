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
		log.Fatal(err)
	}

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGABRT, syscall.SIGINT)
	transmissionClient, err := torrent.NewTransmissionClient(appConfig.Torrent.Transmission)
	if err != nil {
		log.Fatal(err)
	}

	torrentService := torrent.NewTorrentService(transmissionClient)
	torrentProcessor := torrent.NewFinishedTorrentProcessor(transmissionClient, appConfig.Paths)
	appContext, cancel := context.WithCancel(context.Background())
	wg := &sync.WaitGroup{}
	finishedTorrentChan := make(chan torrent.AddedTorrent, 3)

	wg.Add(1)
	go func() {
		defer wg.Done()
		api.StartServer(appContext, appConfig, torrentService)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		torrentService.StartPolling(appContext, finishedTorrentChan, appConfig.Torrent.PollingInterval)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		torrentProcessor.HandleFinishedTorrent(appContext, finishedTorrentChan)
	}()
	sig := <-signalChan
	log.Printf("Signal received: %s", sig)
	cancel()
	wg.Wait()
}

func parseFlags() {
	flag.StringVar(&configFilePath, "configFilePath", "./config.yml", "Path to config file")
	flag.Parse()
}
