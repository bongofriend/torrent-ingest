package torrent

import (
	"context"
	"fmt"
	"log"
	"path/filepath"
	"sync"
	"time"

	"github.com/bongofriend/torrent-ingest/config"
	"github.com/bongofriend/torrent-ingest/models"
	cp "github.com/otiai10/copy"
)

const (
	concurrentJobLimit uint8 = 3
)

type FinishedTorrentPostProcessor interface {
	Start(ctx context.Context, interval time.Duration)
}

type finishedTorrentPostProcessor struct {
	client            TransmissionClient
	pathConfig        config.PathConfig
	concurrentJobChan chan any
}

func NewFinishedTorrentProcessor(t TransmissionClient, d config.PathConfig) FinishedTorrentPostProcessor {
	return finishedTorrentPostProcessor{
		client:            t,
		pathConfig:        d,
		concurrentJobChan: make(chan any, concurrentJobLimit),
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
			f.concurrentJobChan <- struct{}{}
			go func() {
				defer func() {
					<-f.concurrentJobChan
				}()
				if err := f.client.RemoveTorrent(ctx, t); err != nil {
					log.Println(err)
					return
				}
				var dest *string
				switch t.Category {
				case models.Audiobook:
					dest = f.pathConfig.Destinations.Audiobooks
				case models.Anime:
					dest = f.pathConfig.Destinations.Anime
				case models.Series:
					dest = f.pathConfig.Destinations.Series
				case models.Movies:
					dest = f.pathConfig.Destinations.Movie
				default:
					log.Printf("Unknown category %s for torrent %s", t.Category, t.Hash)
					return
				}

				if err := f.copy(t, dest); err != nil {
					log.Println(err)
					return
				}
			}()
		}
	}
}

func (f finishedTorrentPostProcessor) copy(t AddedTorrent, dest *string) error {
	if dest == nil {
		return fmt.Errorf("no destination defined for category %s", t.Category)
	}
	for _, fi := range t.FileNames {
		srcPath := filepath.Join(f.pathConfig.DownloadBasePath, fi)
		destPath := filepath.Join(*dest, fi)
		if err := cp.Copy(srcPath, destPath); err != nil {
			return err
		}
	}
	return nil
}
