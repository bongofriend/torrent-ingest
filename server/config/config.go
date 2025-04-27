package config

import (
	"os"

	"github.com/goccy/go-yaml"
)

type AppConfig struct {
	Server  ServerConfig  `yaml:"server"`
	Torrent TorrentConfig `yaml:"torrent"`
}

type ServerConfig struct {
	Port     uint   `yaml:"port"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

type TorrentConfig struct {
	Transmission TransmissionConfig `yaml:"transmission"`
}

type TransmissionConfig struct {
	Url      string `yaml:"url"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
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
	return config, nil
}
