package torrent

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
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
	Hash      string
	FileNames []string
	Category  models.MediaCategory
}

type TransmissionClient interface {
	AddMagnetLink(context context.Context, category models.MediaCategory, magnetLinkUrl string) (AddedTorrent, error)
	GetAllTorrents() ([]AddedTorrent, error)
}

type transmissionClient struct {
	client *transmissionrpc.Client
}

// GetAllTorrentMagnetLinks implements TransmissionClient.
func (t transmissionClient) GetAllTorrents() ([]AddedTorrent, error) {
	allTorrents, err := t.client.TorrentGetAll(context.Background())
	if err != nil {
		return nil, err
	}
	torrents := make([]AddedTorrent, len(allTorrents))
	for i, t := range allTorrents {
		if t.MagnetLink == nil || t.Files == nil || (t.IsFinished != nil && *t.IsFinished) {
			continue
		}
		id, err := getId(bytes.NewBufferString(*t.MagnetLink))
		if err != nil {
			return nil, err
		}
		category, err := decodeCategoryFromLabels(t.Labels)
		if err != nil {
			continue
		}
		filenames := make([]string, len(t.Files))
		for j, f := range t.Files {
			filenames[j] = f.Name
		}
		torrents[i] = AddedTorrent{
			Hash:      id,
			FileNames: filenames,
			Category:  models.MediaCategory(category),
		}
	}
	return torrents, nil
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

// AddMagnetLink implements TransmissionClient.
func (t transmissionClient) AddMagnetLink(context context.Context, category models.MediaCategory, magnetLinkUrl string) (AddedTorrent, error) {
	id, err := getId(bytes.NewBufferString(magnetLinkUrl))
	if err != nil {
		return AddedTorrent{}, err
	}
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
		Hash:      id,
		FileNames: filenames,
		Category:  category,
	}, err
}

func getId(data io.Reader) (string, error) {
	buf, err := io.ReadAll(data)
	if err != nil {
		return "", err
	}
	hasher := md5.New()
	if _, err := hasher.Write(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(hasher.Sum(nil)), nil
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
