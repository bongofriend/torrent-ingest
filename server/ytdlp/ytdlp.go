package ytdlp

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/bongofriend/torrent-ingest/config"
	"github.com/bongofriend/torrent-ingest/models"
	"github.com/lrstanley/go-ytdlp"
	cp "github.com/otiai10/copy"
)

var (
	ErrNotEnqueued error = errors.New("could not enqueue download")
)

const (
	maxParallelDownloadLimit int           = 5
	maxDownloadEnqueTimeout  time.Duration = 3 * time.Second
)

type AddDownloadRequest struct {
	Url      string
	UrlType  models.YoutubeUrlType
	Category models.MediaCategory
}

type YtdlpDownloadService interface {
	QueueDownload(ctx context.Context, request AddDownloadRequest) error
}

type YtdlpService interface {
	YtdlpDownloadService
	Start(ctx context.Context)
}

type ytdlpCommandFunc func() *ytdlp.Command

type ytdlpService struct {
	jobChan       chan AddDownloadRequest
	ytdlpCommands map[models.MediaCategory]ytdlpCommandFunc
	pathConfig    config.PathConfig
}

func configureForMusic() *ytdlp.Command {
	return ytdlp.New().
		PrintJSON().
		ExtractAudio().
		AudioFormat("mp3").
		NoKeepVideo().
		NoOverwrites().
		EmbedThumbnail().
		EmbedMetadata().
		Continue().
		ProgressFunc(100*time.Millisecond, func(prog ytdlp.ProgressUpdate) {
			fmt.Printf( //nolint:forbidigo
				"%s @ %s [eta: %s] :: %s\n",
				prog.Status,
				prog.PercentString(),
				prog.ETA(),
				prog.Filename,
			)
		})
}

func configureForVideo() *ytdlp.Command {
	return ytdlp.New().
		PrintJSON().
		Format("mp4").
		NoOverwrites().
		Continue().
		EmbedChapters().
		EmbedMetadata().
		EmbedThumbnail().
		ProgressFunc(100*time.Millisecond, func(prog ytdlp.ProgressUpdate) {
			fmt.Printf( //nolint:forbidigo
				"%s @ %s [eta: %s] :: %s\n",
				prog.Status,
				prog.PercentString(),
				prog.ETA(),
				prog.Filename,
			)
		})
}

func NewYtlDlpService(pathConfig config.PathConfig) YtdlpService {
	return ytdlpService{
		pathConfig: pathConfig,
		jobChan:    make(chan AddDownloadRequest, maxParallelDownloadLimit),
		ytdlpCommands: map[models.MediaCategory]ytdlpCommandFunc{
			models.Music:  configureForMusic,
			models.Series: configureForVideo,
		},
	}
}

// QueueDownload implements YtdlpyService.
func (y ytdlpService) QueueDownload(ctx context.Context, request AddDownloadRequest) error {
	// Create a context that cancels itself after some time
	ctxWithTimeout, cancel := context.WithTimeout(ctx, maxDownloadEnqueTimeout)
	defer cancel()
	select {
	case <-ctxWithTimeout.Done():
		return ErrNotEnqueued
	case y.jobChan <- request:
		return nil
	}
}

// Start implements YtdlpService.
func (y ytdlpService) Start(ctx context.Context) {
	defer func() {
		close(y.jobChan)
	}()
	for {
		select {
		case <-ctx.Done():
			return
		case job, ok := <-y.jobChan:
			if !ok {
				return
			}
			if err := y.handleDownload(ctx, job); err != nil {
				log.Println(err)
				continue
			}
		}
	}
}

func (y ytdlpService) handleDownload(ctx context.Context, job AddDownloadRequest) error {
	log.Printf("Downloading Yotube URL %s as %s for media category %s", job.Url, job.UrlType, job.Category)
	workingDir, err := os.MkdirTemp("", "ytldlp*")
	if err != nil {
		return err
	}
	defer func() {
		os.RemoveAll(workingDir)
	}()
	commandFunc, ok := y.ytdlpCommands[job.Category]
	if !ok {
		return fmt.Errorf("media catefory %s not supported by ytdlp", job.Category)
	}
	ytdlpCmd := commandFunc().
		Paths(workingDir)
	if job.UrlType != models.Playlist {
		ytdlpCmd = ytdlpCmd.NoPlaylist()
	}
	_, err = ytdlpCmd.Run(ctx, job.Url)
	if err != nil {
		return err
	}
	y.copyDownloads(workingDir, job)
	log.Printf("Finished downloading Yotube URL %s as %s for media category %s", job.Url, job.UrlType, job.Category)
	return nil
}

func (y ytdlpService) copyDownloads(downloadPath string, job AddDownloadRequest) {
	var dest string
	switch job.Category {
	case models.Audiobook:
		dest = y.pathConfig.Destinations.Audiobooks
	case models.Anime:
		dest = y.pathConfig.Destinations.Anime
	case models.Series:
		dest = y.pathConfig.Destinations.Series
	case models.Movies:
		dest = y.pathConfig.Destinations.Movie
	case models.Music:
		dest = y.pathConfig.Destinations.Music
	default:
		log.Printf("Unknown category %s for download %s", job.Category, job.Url)
		return
	}
	if len(dest) == 0 {
		log.Printf("No destination configured for category %s, skipping download %s", job.Category, job.Url)
		return
	}
	if err := cp.Copy(downloadPath, dest); err != nil {
		log.Print(err)
	}
}
