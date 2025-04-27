package torrent

import (
	"context"
	"errors"
	"sync"

	"github.com/bongofriend/torrent-ingest/config"
	"github.com/bongofriend/torrent-ingest/models"
)

var (
	ErrTorrentAlreadyAdded   error = errors.New("torrent is already added")
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
	Init() error
	AddMagnet(ctx context.Context, req AddMagnetLinkRequest) error
}

type torrentService struct {
	addedFilesMutex *sync.RWMutex
	addedFiles      map[string]AddedTorrent

	transmissionClient TransmissionClient
}

func NewTorrentService(config config.TorrentConfig) (TorrentService, error) {
	client, err := NewTransmissionClient(config.Transmission)
	if err != nil {
		return nil, err
	}
	return &torrentService{
		addedFilesMutex:    &sync.RWMutex{},
		addedFiles:         map[string]AddedTorrent{},
		transmissionClient: client,
	}, nil
}

// Init implements TorrentService.
func (t *torrentService) Init() error {
	allTorrents, err := t.transmissionClient.GetAllTorrents()
	if err != nil {
		return err
	}
	for _, to := range allTorrents {
		t.addedFiles[to.Hash] = to
	}
	return nil
}

// AddMagnet implements TorrentService.
func (t *torrentService) AddMagnet(ctx context.Context, req AddMagnetLinkRequest) error {
	tor, err := t.transmissionClient.AddMagnetLink(ctx, req.Category, req.MagnetLink)
	if err != nil {
		return err
	}
	t.addedFilesMutex.RLock()
	if _, ok := t.addedFiles[tor.Hash]; ok {
		t.addedFilesMutex.RUnlock()
		return ErrTorrentAlreadyAdded
	}
	t.addedFilesMutex.RUnlock()
	t.addedFilesMutex.Lock()
	t.addedFiles[tor.Hash] = tor
	t.addedFilesMutex.Unlock()
	return nil
}
