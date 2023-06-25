package config

import (
	"log"

	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	ClientID string `envconfig:"CLIENT_ID" required:"true"`
	UserID   string `envconfig:"USER_ID" default:"fmcyamwe" required:"true"`
}

func FromEnvironment() (Config, error) {
	var cfg Config

	err := envconfig.Process("", &cfg)

	if err != nil {
		log.Println("ooh oh cant populate Config")
		return Config{}, err
	}

	return cfg, nil
}
