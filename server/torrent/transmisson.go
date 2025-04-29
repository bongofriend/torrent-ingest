package torrent

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/bongofriend/torrent-ingest/config"
	"github.com/bongofriend/torrent-ingest/models"
	"github.com/hekmon/transmissionrpc/v3"
)

const (
	categoryLabelPrefix string = "Category"
)

var (
	errCategoryNotFound error = errors.New("category not found in torrent labels")
)

type AddedTorrent struct {
	Id        int64
	Hash      string
	FileNames []string
	Category  models.MediaCategory
}

type TransmissionClient interface {
	AddMagnetLink(context context.Context, category models.MediaCategory, magnetLinkUrl string) (AddedTorrent, error)
	GetAllFinishedTorrents(ctx context.Context) ([]AddedTorrent, error)
	RemoveTorrent(ctx context.Context, torrent AddedTorrent) error
}

type transmissionClient struct {
	client *transmissionrpc.Client
}

func NewTransmissionClient(transmissionConfig config.TransmissionConfig) (TransmissionClient, error) {
	transmissionUrl, err := url.Parse(transmissionConfig.Url)
	if err != nil {
		return nil, err
	}
	transmissionUrl.User = url.UserPassword(transmissionConfig.Username, transmissionConfig.Password)
	client, err := transmissionrpc.New(transmissionUrl, nil)
	if err != nil {
		return nil, err
	}
	return transmissionClient{
		client: client,
	}, err
}

// GetAllTorrentMagnetLinks implements TransmissionClient.
func (t transmissionClient) GetAllFinishedTorrents(ctx context.Context) ([]AddedTorrent, error) {
	allTorrents, err := t.client.TorrentGetAll(ctx)
	if err != nil {
		return nil, err
	}
	torrents := []AddedTorrent{}
	for _, t := range allTorrents {
		if t.MagnetLink == nil || t.Files == nil || (t.PercentDone != nil && *t.PercentDone < 1.0) {
			continue
		}
		category, err := decodeCategoryFromLabels(t.Labels)
		if err != nil {
			continue
		}
		filenames := make([]string, len(t.Files))
		for j, f := range t.Files {
			filenames[j] = f.Name
		}
		torrents = append(torrents, AddedTorrent{
			Id:        *t.ID,
			Hash:      *t.HashString,
			FileNames: filenames,
			Category:  models.MediaCategory(category),
		})
	}
	return torrents, nil
}

// AddMagnetLink implements TransmissionClient.
func (t transmissionClient) AddMagnetLink(context context.Context, category models.MediaCategory, magnetLinkUrl string) (AddedTorrent, error) {
	payload := transmissionrpc.TorrentAddPayload{
		Filename: &magnetLinkUrl,
		Labels:   []string{encodeCatgeoryAsLabel(category)},
	}
	to, err := t.client.TorrentAdd(context, payload)
	if err != nil {
		return AddedTorrent{}, err
	}
	filenames := make([]string, len(to.Files))
	for j, f := range to.Files {
		filenames[j] = f.Name
	}
	return AddedTorrent{
		Id:        *to.ID,
		Hash:      *to.HashString,
		FileNames: filenames,
		Category:  category,
	}, err
}

func encodeCatgeoryAsLabel(category models.MediaCategory) string {
	return fmt.Sprintf("%s:%s", categoryLabelPrefix, category)
}
func decodeCategoryFromLabels(labels []string) (string, error) {
	for _, s := range labels {
		if strings.HasPrefix(s, categoryLabelPrefix) {
			return strings.TrimPrefix(s, fmt.Sprintf("%s:", categoryLabelPrefix)), nil
		}
	}
	return "", errCategoryNotFound
}

// StopTorrent implements TransmissionClient.
func (t transmissionClient) RemoveTorrent(ctx context.Context, torrent AddedTorrent) error {
	return t.client.TorrentRemove(ctx, transmissionrpc.TorrentRemovePayload{
		IDs:             []int64{torrent.Id},
		DeleteLocalData: false,
	})
}
