package proto

import (
	"os"
)

type Config struct {
	DeviceName  string
	DeviceId    string
	Domain      string
	Server      string
	Certificate string
	Key         string
	Authority   string
}

func (cfg *Config) LoadFromEnv() {
	if cfg.Server == "" {
		cfg.Server = os.Getenv("STARK_HOST")
	}
	if cfg.Certificate == "" {
		cfg.Certificate = os.Getenv("STARK_CERT")
	}
	if cfg.Key == "" {
		cfg.Key = os.Getenv("STARK_KEY")
	}
	if cfg.Authority == "" {
		cfg.Authority = os.Getenv("STARK_CA")
	}
}
