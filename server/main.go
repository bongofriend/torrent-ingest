package main

import (
	"context"
	"errors"
	"flag"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

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

	torrentProcessor := torrent.NewFinishedTorrentProcessor(transmissionClient, appConfig.Paths)
	appContext, cancel := context.WithCancel(context.Background())
	wg := &sync.WaitGroup{}

	wg.Add(1)
	go func() {
		defer wg.Done()
		api.StartServer(appContext, appConfig, transmissionClient)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		torrentProcessor.Start(appContext, appConfig.Torrent.PollingInterval)
	}()
	sig := <-signalChan
	log.Printf("Signal received: %s", sig)
	cancel()

	shutdownContext, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	go func() {
		<-shutdownContext.Done()
		err := shutdownContext.Err()
		if errors.Is(err, context.DeadlineExceeded) {
			log.Panic("Timeout for shutdown reached")
		}
	}()

	wg.Wait()
	cancel()
}

func parseFlags() {
	flag.StringVar(&configFilePath, "configFilePath", "./config.yml", "Path to config file")
	flag.Parse()
}
