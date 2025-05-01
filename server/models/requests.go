package models

import validation "github.com/go-ozzo/ozzo-validation"

type MediaCategory string

const (
	Audiobook MediaCategory = "audiobook"
	Anime     MediaCategory = "anime"
)

func (m MediaCategory) Validate() error {
	return validation.Validate(string(m), validation.Required, validation.In(string(Audiobook), string(Anime)))
}
