package api

import (
	"context"
	"errors"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/bongofriend/torrent-ingest/config"
	"github.com/bongofriend/torrent-ingest/torrent"
)

func Run(ctx context.Context, appConfig config.AppConfig) {

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
		startServer(appContext, appConfig, transmissionClient)
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
