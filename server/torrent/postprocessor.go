package torrent

import (
	"context"
	"log"
	"path/filepath"
	"sync"
	"time"

	"github.com/bongofriend/torrent-ingest/config"
	"github.com/bongofriend/torrent-ingest/models"
	cp "github.com/otiai10/copy"
)

type FinishedTorrentPostProcessor interface {
	Start(ctx context.Context, interval time.Duration)
}

type finishedTorrentPostProcessor struct {
	client     TransmissionClient
	pathConfig config.PathConfig
}

func NewFinishedTorrentProcessor(t TransmissionClient, d config.PathConfig) FinishedTorrentPostProcessor {
	return finishedTorrentPostProcessor{
		client:     t,
		pathConfig: d,
	}
}

func (f finishedTorrentPostProcessor) Start(ctx context.Context, interval time.Duration) {
	wg := &sync.WaitGroup{}
	finishedTorrentChan := make(chan AddedTorrent, 3)

	wg.Add(1)
	go func() {
		defer wg.Done()
		f.poll(ctx, interval, finishedTorrentChan)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		f.handleFinishedTorrent(ctx, finishedTorrentChan)
	}()

	wg.Wait()
}

func (f finishedTorrentPostProcessor) poll(ctx context.Context, interval time.Duration, finishedTorrentsChan chan<- AddedTorrent) {
	ticker := time.NewTicker(interval)
	for {
		select {
		case <-ctx.Done():
			log.Println("Polling for finished torrents stopped")
			return
		case <-ticker.C:
			torrents, err := f.client.GetAllFinishedTorrents(ctx)
			if err != nil {
				log.Println(err)
				continue
			}
			for _, to := range torrents {
				finishedTorrentsChan <- to
			}
		}
	}
}

func (f finishedTorrentPostProcessor) handleFinishedTorrent(ctx context.Context, finishedTorrents <-chan AddedTorrent) {
	for {
		select {
		case <-ctx.Done():
			log.Println("Processing of finished torrents stopped")
			return
		case t := <-finishedTorrents:
			if err := f.client.RemoveTorrent(ctx, t); err != nil {
				log.Println(err)
				continue
			}
			var dest string
			switch t.Category {
			case models.Audiobook:
				dest = f.pathConfig.Destinations.Audiobooks
			case models.Anime:
				dest = f.pathConfig.Destinations.Anime
			}
			if err := f.copy(t, dest); err != nil {
				log.Println(err)
				continue
			}
		}
	}
}

// Create shard docker volume between torrent-ingest, transmission and audiobookshelf
func (f finishedTorrentPostProcessor) copy(t AddedTorrent, dest string) error {
	for _, fi := range t.FileNames {
		srcPath := filepath.Join(f.pathConfig.DownloadBasePath, fi)
		destPath := filepath.Join(dest, fi)
		if err := cp.Copy(srcPath, destPath); err != nil {
			return err
		}
	}
	return nil
}
