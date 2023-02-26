package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

type Config struct {
	ConfigFile string
	IssuerURL  string
	AuthURL    string

	Listeners struct {
		Frontend string `yaml:"frontend"`
		Backend  string `yaml:"backend"`
	}

	ContentDir string `yaml:"content_dir"`

	Https struct {
		CaCertFile string `yaml:"ca_cert_file"`
		KeyFile    string `yaml:"key_file"`
		CertFile   string `yaml:"cert_file"`
	}

	Clients []ClientConfig `yaml:"clients"`

	UserDb UserDb `yaml:"user_db"`
}

type ClientConfig struct {
	Id           string   `yaml:"id"`
	Secret       string   `yaml:"secret"`
	RedirectURLs []string `yaml:"redirect_urls"`
}

type UserDb struct {
	Path string `yaml:"path"`
}

func Load(file string) (*Config, error) {
	// load the config
	fh, err := os.Open(file)
	if err != nil {
		return nil, fmt.Errorf("failed to open configuration file: %w", err)
	}
	defer fh.Close()

	decoder := yaml.NewDecoder(fh)

	var cfg Config
	err = decoder.Decode(&cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to read configuration file: %w", err)
	}

	// set the derived variables
	cfg.ConfigFile = file

	configdir := filepath.Dir(file)
	if !strings.HasPrefix(cfg.Https.CaCertFile, "/") {
		cfg.Https.CaCertFile = filepath.Join(configdir, cfg.Https.CaCertFile)
	}
	if !strings.HasPrefix(cfg.Https.KeyFile, "/") {
		cfg.Https.KeyFile = filepath.Join(configdir, cfg.Https.KeyFile)
	}
	if !strings.HasPrefix(cfg.Https.CertFile, "/") {
		cfg.Https.CertFile = filepath.Join(configdir, cfg.Https.CertFile)
	}

	cfg.Listeners.Backend = strings.TrimRight(cfg.Listeners.Backend, "/")
	cfg.Listeners.Frontend = strings.TrimRight(cfg.Listeners.Frontend, "/")
	cfg.IssuerURL = cfg.Listeners.Backend
	cfg.AuthURL = strings.Join([]string{cfg.Listeners.Frontend, "auth"}, "/")

	return &cfg, nil
}