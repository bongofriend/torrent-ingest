package api

import (
	"encoding/json"
	"io"
	"log"
	"net/http"

	"github.com/bongofriend/torrent-ingest/models"
	"github.com/bongofriend/torrent-ingest/torrent"
	validation "github.com/go-ozzo/ozzo-validation"
)

const (
	uploadMaxFileSize       int64  = 1 * 1024 * 1024 //1 MB file limit
	fileUploadFormName      string = "torrent"
	mediaCategoryQueryParam string = "category"
)

type magnetLinkRequestBody struct {
	Category   models.MediaCategory `json:"category"`
	MagnetLink string               `json:"magnetLink"`
}

func (m magnetLinkRequestBody) Validate() error {
	return validation.ValidateStruct(&m,
		validation.Field(&m.Category),
		validation.Field(&m.MagnetLink, validation.Required),
	)
}

type torrentFileRequestBody struct {
	Category           models.MediaCategory
	TorrentFileContent []byte
}

func (t torrentFileRequestBody) Validate() error {
	return validation.ValidateStruct(&t,
		validation.Field(&t.Category),
		validation.Field(&t.TorrentFileContent, validation.NilOrNotEmpty),
	)
}

func registerEndpoints(mux *http.ServeMux, transmissionClient torrent.TransmissionClient) {
	mux.HandleFunc("POST /torrent/magnetlink", handleMagnetLink(transmissionClient))
	mux.HandleFunc("POST /torrent/file", handleTorrentFile(transmissionClient))
}

func handleMagnetLink(transmissionClient torrent.TransmissionClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var requestBody magnetLinkRequestBody
		if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
			log.Println(err)
			http.Error(w, internalServerErrorMessage, http.StatusInternalServerError)
			return
		}
		if err := requestBody.Validate(); err != nil {
			log.Println(err)
			http.Error(w, badRequestMessage, http.StatusBadRequest)
			return
		}

		if _, err := transmissionClient.AddMagnetLink(r.Context(), torrent.AddMagnetLinkRequest{
			Category:   requestBody.Category,
			MagnetLink: requestBody.MagnetLink,
		}); err != nil {
			log.Println(err)
			http.Error(w, internalServerErrorMessage, http.StatusInternalServerError)
		}
	}
}

func handleTorrentFile(transmissionClient torrent.TransmissionClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		queryValue := r.URL.Query().Get(mediaCategoryQueryParam)
		if len(queryValue) == 0 {
			http.Error(w, badRequestMessage, http.StatusBadRequest)
			return
		}

		r.Body = http.MaxBytesReader(w, r.Body, uploadMaxFileSize)
		file, _, err := r.FormFile(fileUploadFormName)
		if err != nil {
			log.Println(err)
			http.Error(w, badRequestMessage, http.StatusRequestEntityTooLarge)
			return
		}

		content, err := io.ReadAll(file)
		if err != nil {
			log.Println(err)
			http.Error(w, internalServerErrorMessage, http.StatusInternalServerError)
			return
		}
		file.Close()

		request := torrentFileRequestBody{
			Category:           models.MediaCategory(queryValue),
			TorrentFileContent: content,
		}

		if err := request.Validate(); err != nil {
			log.Println(err)
			http.Error(w, badRequestMessage, http.StatusBadRequest)
			return
		}

		if _, err := transmissionClient.AddTorrentFile(r.Context(), torrent.AddTorrentFileRequest{
			Category:           request.Category,
			TorrentFileContent: request.TorrentFileContent,
		}); err != nil {
			log.Println(err)
			http.Error(w, internalServerErrorMessage, http.StatusInternalServerError)
			return
		}
	}
}
