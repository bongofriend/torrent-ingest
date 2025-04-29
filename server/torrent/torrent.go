package torrent

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/bongofriend/torrent-ingest/models"
)

var (
	ErrUnknownTorrentRequest error = errors.New("request to add torrent not supported")
)

type AddMagnetLinkRequest struct {
	Category   models.MediaCategory
	MagnetLink string
}

func (a AddMagnetLinkRequest) MediaCategory() models.MediaCategory {
	return a.Category
}

type TorrentService interface {
	AddMagnet(ctx context.Context, req AddMagnetLinkRequest) error
	StartPolling(ctx context.Context, finishedTorrents chan<- AddedTorrent, interval time.Duration)
}

type torrentService struct {
	transmissionClient TransmissionClient
}

func NewTorrentService(t TransmissionClient) TorrentService {
	return &torrentService{
		transmissionClient: t,
	}
}

// AddMagnet implements TorrentService.
func (t *torrentService) AddMagnet(ctx context.Context, req AddMagnetLinkRequest) error {
	_, err := t.transmissionClient.AddMagnetLink(ctx, req.Category, req.MagnetLink)
	return err
}

// StartPolling implements TorrentService.
func (t *torrentService) StartPolling(ctx context.Context, finishedTorrentsChan chan<- AddedTorrent, interval time.Duration) {
	ticker := time.NewTicker(interval)
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			torrents, err := t.transmissionClient.GetAllFinishedTorrents(ctx)
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
