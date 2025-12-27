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
	"github.com/bongofriend/torrent-ingest/ytdlp"
)

func Run(ctx context.Context, appConfig config.AppConfig) {

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGABRT, syscall.SIGINT)
	transmissionClient, err := torrent.NewTransmissionClient(appConfig.Torrent.Transmission)
	if err != nil {
		log.Fatal(err)
	}

	printConfig(appConfig)

	torrentProcessor := torrent.NewFinishedTorrentProcessor(transmissionClient, appConfig.Paths)
	appContext, cancel := context.WithCancel(context.Background())
	wg := &sync.WaitGroup{}

	ytdlpService := ytdlp.NewYtlDlpService(appConfig.Paths)

	wg.Add(1)
	go func() {
		defer wg.Done()
		torrentProcessor.Start(appContext, appConfig.Torrent.PollingInterval)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		ytdlpService.Start(ctx)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		startServer(appContext, appConfig, transmissionClient, ytdlpService)
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

func printConfig(appConfig config.AppConfig) {
	log.Println("Application configuration:")
	log.Printf(" - Server port: %d", appConfig.Server.Port)
	log.Printf(" - Torrent polling interval: %s", appConfig.Torrent.PollingInterval)
	log.Printf(" - Transmission URL: %s", appConfig.Torrent.Transmission.Url)
	log.Printf(" - Paths:")
	if appConfig.Paths.Destinations.Audiobooks != "" {
		log.Printf("   - Audiobooks: %s", appConfig.Paths.Destinations.Audiobooks)
	}
	if appConfig.Paths.Destinations.Anime != "" {
		log.Printf("   - Anime: %s", appConfig.Paths.Destinations.Anime)
	}
	if appConfig.Paths.Destinations.Movie != "" {
		log.Printf("   - Movie: %s", appConfig.Paths.Destinations.Movie)
	}
	if appConfig.Paths.Destinations.Series != "" {
		log.Printf("   - Series: %s", appConfig.Paths.Destinations.Series)
	}
	if appConfig.Paths.Destinations.Music != "" {
		log.Printf("   - Music: %s", appConfig.Paths.Destinations.Music)
	}
}
