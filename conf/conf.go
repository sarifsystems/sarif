package conf

import (
	"encoding/json"
	"os"
	"path"
)

type Config struct {
	Db struct {
		Driver string
		Source string
	}
	Proto struct {
		Domain      string
		Server      string
		Certificate string
		Key         string
		Authority   string
	}
}

func GetDefaults() Config {
	var cfg Config
	cfg.Db.Driver = "sqlite3"
	cfg.Db.Source = GetDefaultDir() + "/stark.db"
	return cfg
}

func Read(file string) (Config, error) {
	if file == "" {
		file = GetDefaultPath()
	}

	var cfg Config
	f, err := os.Open(file)
	if err != nil {
		return cfg, err
	}
	defer f.Close()
	dec := json.NewDecoder(f)
	err = dec.Decode(&cfg)
	return cfg, err
}

func ReadDefault() (Config, error) {
	cfg, err := Read(GetDefaultPath())
	if err != nil && os.IsNotExist(err) {
		cfg = GetDefaults()
		err = Write(GetDefaultPath(), cfg)
	}
	return cfg, err
}

func Write(file string, cfg Config) error {
	if file == "" {
		file = GetDefaultPath()
	}

	if err := os.MkdirAll(path.Dir(file), 0700); err != nil {
		return err
	}

	f, err := os.Create(file)
	if err != nil {
		return err
	}
	defer f.Close()

	encoded, err := json.MarshalIndent(&cfg, "", "\t")
	if err != nil {
		return err
	}
	_, err = f.Write(encoded)
	return err
}

func GetDefaultPath() string {
	return GetDefaultDir() + "/config.json"
}

func GetDefaultDir() string {
	path := os.Getenv("XDG_CONFIG_HOME")
	if path != "" {
		return path + "/stark"
	}

	home := os.Getenv("HOME")
	if home != "" {
		return home + "/.config/stark"
	}

	return "."
}
