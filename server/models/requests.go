package models

import validation "github.com/go-ozzo/ozzo-validation"

type MediaCategory string

const (
	Audiobook MediaCategory = "audiobook"
	Anime     MediaCategory = "anime"
	Series    MediaCategory = "series"
	Movies    MediaCategory = "movies"
	Music     MediaCategory = "music"
)

func (m MediaCategory) Validate() error {
	return validation.Validate(string(m), validation.Required, validation.In(string(Audiobook), string(Anime), string(Series), string(Movies), string(Music)))
}

type YoutubeUrlType string

const (
	Video    YoutubeUrlType = "video"
	Playlist YoutubeUrlType = "playlist"
)

func (y YoutubeUrlType) Validate() error {
	return validation.Validate(string(y), validation.Required, validation.In(string(Video), string(Playlist)))
}
