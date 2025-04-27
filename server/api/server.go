package api

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/bongofriend/torrent-ingest/config"
	"github.com/bongofriend/torrent-ingest/torrent"
)

func StartServer(appContext context.Context, wg *sync.WaitGroup, appConfig config.AppConfig, torrentService torrent.TorrentService) {
	defer wg.Done()
	apiMux := http.NewServeMux()
	registerEndpoints(apiMux, torrentService)

	middleware := applyMiddleware(logging(), auth(appConfig.Server))
	server := &http.Server{
		Addr:    fmt.Sprintf("0.0.0.0:%d", appConfig.Server.Port),
		Handler: middleware(apiMux),
	}

	go func() {
		log.Printf("Server listening on port %d", appConfig.Server.Port)
		if err := server.ListenAndServe(); err != nil {
			panic(err)
		}
	}()

	<-appContext.Done()
	shutdownContext, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := server.Shutdown(shutdownContext); err != nil {
		panic(err)
	}
}
