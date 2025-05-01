package config

import (
	"os"
	"time"

	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/go-ozzo/ozzo-validation/is"
	"github.com/goccy/go-yaml"
)

type AppConfig struct {
	Server  ServerConfig  `yaml:"server"`
	Torrent TorrentConfig `yaml:"torrent"`
	Paths   PathConfig    `yaml:"paths"`
}

func (a AppConfig) Validate() error {
	return validation.ValidateStruct(&a,
		validation.Field(&a.Server),
		validation.Field(&a.Torrent),
		validation.Field(&a.Paths),
	)
}

type ServerConfig struct {
	Port     int    `yaml:"port"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

func (s ServerConfig) Validate() error {
	return validation.ValidateStruct(&s,
		validation.Field(&s.Port, validation.Min(0)),
		validation.Field(&s.Username, validation.NilOrNotEmpty),
		validation.Field(&s.Password, validation.NilOrNotEmpty, validation.Length(10, 0)),
	)
}

type TorrentConfig struct {
	PollingInterval time.Duration      `yaml:"polling_interval"`
	Transmission    TransmissionConfig `yaml:"transmission"`
}

func (t TorrentConfig) Validate() error {
	return validation.ValidateStruct(&t,
		validation.Field(&t.PollingInterval, validation.Required),
		validation.Field(&t.Transmission),
	)
}

type PathConfig struct {
	DownloadBasePath string             `yaml:"download_base_path"`
	Destinations     DestionationConfig `yaml:"destinations"`
}

func (p PathConfig) Validate() error {
	return validation.ValidateStruct(&p,
		validation.Field(&p.DownloadBasePath, validation.NilOrNotEmpty),
		validation.Field(&p.Destinations),
	)
}

type DestionationConfig struct {
	Audiobooks string `yaml:"audiobooks"`
	Anime      string `yaml:"anime"`
}

func (d DestionationConfig) Validate() error {
	return validation.ValidateStruct(&d,
		validation.Field(&d.Audiobooks, validation.NilOrNotEmpty),
		validation.Field(&d.Anime, validation.NilOrNotEmpty),
	)
}

type TransmissionConfig struct {
	Url      string `yaml:"url"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

func (t TransmissionConfig) Validate() error {
	return validation.ValidateStruct(&t,
		validation.Field(&t.Url, is.URL),
		validation.Field(&t.Username, validation.NilOrNotEmpty),
		validation.Field(&t.Password, validation.NilOrNotEmpty),
	)
}

func LoadConfig(configFilePath string) (AppConfig, error) {
	configFile, err := os.Open(configFilePath)
	if err != nil {
		return AppConfig{}, err
	}
	defer configFile.Close()

	var config AppConfig
	if err = yaml.NewDecoder(configFile).Decode(&config); err != nil {
		return AppConfig{}, err
	}
	if err := config.Validate(); err != nil {
		return AppConfig{}, err
	}
	return config, nil
}
