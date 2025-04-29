package torrent

import (
	"context"
	"log"
	"path/filepath"

	"github.com/bongofriend/torrent-ingest/config"
	"github.com/bongofriend/torrent-ingest/models"
	cp "github.com/otiai10/copy"
)

type FinishedTorrentPostProcessor interface {
	HandleFinishedTorrent(ctx context.Context, finishedTorrents <-chan AddedTorrent)
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

// HandleFinishedTorrent implements FinishedTorrentProcessor.
func (f finishedTorrentPostProcessor) HandleFinishedTorrent(ctx context.Context, finishedTorrents <-chan AddedTorrent) {
	for {
		select {
		case <-ctx.Done():
			return
		case t := <-finishedTorrents:
			if err := f.client.RemoveTorrent(ctx, t); err != nil {
				log.Println(err)
				continue
			}
			switch t.Category {
			case models.Audiobook:
				if err := f.handleAudiobook(t); err != nil {
					log.Println(err)
					continue
				}
			}
		}
	}
}

// Create shard docker volume between torrent-ingest, transmission and audiobookshelf
func (f finishedTorrentPostProcessor) handleAudiobook(t AddedTorrent) error {
	for _, fi := range t.FileNames {
		srcPath := filepath.Join(f.pathConfig.DownloadBasePath, fi)
		destPath := filepath.Join(f.pathConfig.Destinations.Audiobooks, fi)
		if err := cp.Copy(srcPath, destPath); err != nil {
			return err
		}
	}
	return nil
}
