package api

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/bongofriend/torrent-ingest/models"
	"github.com/bongofriend/torrent-ingest/torrent"
	validation "github.com/go-ozzo/ozzo-validation"
)

type magnetLinkRequestBody struct {
	Category   models.MediaCategory `json:"category"`
	MagnetLink string               `json:"magnetLink"`
}

func (m magnetLinkRequestBody) Validate() error {
	return validation.ValidateStruct(&m,
		validation.Field(&m.Category, validation.Required, validation.In(models.Audiobook)),
		validation.Field(&m.MagnetLink, validation.Required),
	)
}

func registerEndpoints(mux *http.ServeMux, torrentService torrent.TorrentService) {
	mux.HandleFunc("POST /torrent/magnetlink", handleMagnetLink(torrentService))
}

func handleMagnetLink(torrentService torrent.TorrentService) http.HandlerFunc {
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

		err := torrentService.AddMagnet(r.Context(), torrent.AddMagnetLinkRequest{
			Category:   requestBody.Category,
			MagnetLink: requestBody.MagnetLink,
		})
		if err != nil {
			log.Println(err)
			http.Error(w, internalServerErrorMessage, http.StatusInternalServerError)
		}
	}
}
