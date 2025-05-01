package api

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/bongofriend/torrent-ingest/config"
	"github.com/bongofriend/torrent-ingest/torrent"
)

func StartServer(appContext context.Context, appConfig config.AppConfig, transmissionClient torrent.TransmissionClient) {
	apiMux := http.NewServeMux()
	registerEndpoints(apiMux, transmissionClient)

	middleware := applyMiddleware(logging(), auth(appConfig.Server))
	server := &http.Server{
		Addr:    fmt.Sprintf("0.0.0.0:%d", appConfig.Server.Port),
		Handler: middleware(apiMux),
	}

	go func() {
		log.Printf("Server listening on port %d", appConfig.Server.Port)
		if err := server.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			log.Fatal(err)
		}
		log.Println("Server stopped")
	}()

	<-appContext.Done()
	shutdownContext, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := server.Shutdown(shutdownContext); err != nil {
		log.Fatal(err)
	}
}
