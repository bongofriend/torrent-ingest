package main

import (
	"context"
	"os"
	"time"

	"github.com/bongofriend/torrent-ingest/api"
	"github.com/bongofriend/torrent-ingest/config"
	"github.com/urfave/cli/v3"
)

const (
	serverPortEnv string = "TORRENT_INGEST_PORT"
	serverUsernameEnv string = "TORRENT_INGEST_USERNAME"
	serverPasswordEnv string = "TORRENT_INGEST_PASSWORD"

	torrnetPollingIntervalEnv string = "TORRENT_INGEST_POLLING_INTERVAL"
	torrentTransmissionUrlEnv string = "TORRENT_INGEST_TRANSMISSION_URL"
	torrentTransmissionUsernameEnv string = "TORRENT_INGEST_TRANSMISSION_USERNAME"
	torrentTransmissionPasswordEnv string = "TORRENT_INGEST_TRANSMISSION_PASSWORD"

	pathsDownloadBasePathEnv string = "TORRENT_INGEST_DOWNLOAD_BASE_PATH"
	pathsAudiobookPaths string = "TORRENT_INGEST_AUDIOBOOK_PATH"
	pathsSeriesPath string = "TORRENT_INGEST_SERIES_PATH"
	pathsMoviesPath string = "TORRENT_INGEST_MOVIES_PATH"
	pathsAnimePath string = "TORRENT_INGEST_ANIME_PATH"
)

func main() {
	ctx := context.Background()
	if err := rootCmd().Run(ctx, os.Args); err != nil {
		panic(err)
	}
}

func rootCmd() *cli.Command {
	var appConfig config.AppConfig
	
	
	return &cli.Command{
		Name: "torrent-ingest",
		Description: "A minimal go backend for handling downloadinf torrents",
		Flags: []cli.Flag{
			&cli.IntFlag{
				Name: "port",
				Usage: "Port to bind application to",
				Destination: &appConfig.Server.Port,
				Value: 81,
				Sources: cli.EnvVars(serverPortEnv),
			},
			&cli.StringFlag{
				Name: "username",
				Usage: "Username to authencticate requests using Basic Auth",
				Destination: &appConfig.Server.Username,
				Required: true,
				Sources: cli.EnvVars(serverUsernameEnv),
			},
			&cli.StringFlag{
				Name: "password",
				Usage: "Password to authenticate requests using Basic Auth",
				Destination: &appConfig.Server.Password,
				Required: true,
				Sources: cli.EnvVars(serverPasswordEnv),
			},
			&cli.DurationFlag{
				Name: "torrent-polling",
				Usage: "Polling interval to check for finished torrents",
				Destination: &appConfig.Torrent.PollingInterval,
				Value: 30 * time.Second,
				Sources: cli.EnvVars(torrnetPollingIntervalEnv),
			},
			&cli.StringFlag{
				Name: "transmission-url",
				Usage: "URL to transmission RPC endpoint",
				Destination: &appConfig.Torrent.Transmission.Url,
				Required: true,
				Sources: cli.EnvVars(torrentTransmissionUrlEnv),
			},
			&cli.StringFlag{
				Name: "transmission-username",
				Usage: "Username for transmission RPC requests",
				Destination: &appConfig.Torrent.Transmission.Username,
				Required: true,
				Sources: cli.EnvVars(torrentTransmissionUsernameEnv),
			},
			&cli.StringFlag{
				Name: "transmission-password",
				Usage: "Password for transmission RPC requests",
				Destination: &appConfig.Torrent.Transmission.Password,
				Required: true,
				Sources: cli.EnvVars(torrentTransmissionPasswordEnv),
			},
			&cli.StringFlag{
				Name: "download-base-path",
				Usage: "Base path for completed torrent downloads",
				Destination: &appConfig.Paths.DownloadBasePath,
				Required: true,
				Sources: cli.EnvVars(pathsDownloadBasePathEnv),
			},
			&cli.StringFlag{
				Name: "audiobook-path",
				Usage: "Path for downloaded audiobooks",
				Destination: appConfig.Paths.Destinations.Audiobooks,
				Sources: cli.EnvVars(pathsAudiobookPaths),
			},
			&cli.StringFlag{
				Name: "movie-path",
				Usage: "Path for downloaded movies",
				Destination: appConfig.Paths.Destinations.Movie,
				Sources: cli.EnvVars(pathsMoviesPath),
			},
			&cli.StringFlag{
				Name: "series-path",
				Usage: "Path for downloaded series",
				Destination: appConfig.Paths.Destinations.Series,
				Sources: cli.EnvVars(pathsMoviesPath),
			},
			&cli.StringFlag{
				Name: "anime-path",
				Usage: "Path for downloaded animes",
				Destination: appConfig.Paths.Destinations.Anime,
				Sources: cli.EnvVars(pathsAnimePath),
			},
		},
		Action: func(ctx context.Context, c *cli.Command) error {
			if err := appConfig.Validate(); err != nil {
				return err
			}
			api.Run(ctx, appConfig)
			return nil
		},
	}
}
