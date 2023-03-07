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

	ClientDir string `yaml:"client_dir"`
	Clients   []*ClientConfig

	UserDb UserDb `yaml:"user_db"`
}

type ClientConfig struct {
	Id           string   `yaml:"id"`
	Secret       string   `yaml:"secret"`
	RedirectURLs []string `yaml:"redirect_urls"`
}

type UserDb struct {
	Type       string `yaml:"type"`
	Path       string `yaml:"path"`
	LdapServer string `yaml:"ldap_server"`
	LdapPort   int    `yaml:"ldap_port"`
	SearchBase string `yaml:"search_base"`
	SearchDn   string `yaml:"search_dn"`
	SearchPw   string `yaml:"search_pw"`
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
	if !strings.HasPrefix(cfg.ClientDir, "/") {
		cfg.ClientDir = filepath.Join(configdir, cfg.ClientDir)
	}

	cfg.Listeners.Backend = strings.TrimRight(cfg.Listeners.Backend, "/")
	cfg.Listeners.Frontend = strings.TrimRight(cfg.Listeners.Frontend, "/")
	cfg.IssuerURL = cfg.Listeners.Backend
	cfg.AuthURL = strings.Join([]string{cfg.Listeners.Frontend, "auth"}, "/")

	// load the client configs
	entries, err := os.ReadDir(cfg.ClientDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read client configs: %w", err)
	}

	for _, e := range entries {
		if !e.Type().IsRegular() {
			continue
		}
		if !strings.HasSuffix(e.Name(), ".yaml") && !strings.HasSuffix(e.Name(), ".yaml") {
			continue
		}

		fpath := filepath.Join(cfg.ClientDir, e.Name())
		fh, err := os.Open(fpath)
		if err != nil {
			return nil, fmt.Errorf("failed to open client configuration file: %w", err)
		}
		defer fh.Close()

		decoder := yaml.NewDecoder(fh)

		var ccfg ClientConfig
		err = decoder.Decode(&ccfg)
		if err != nil {
			return nil, fmt.Errorf("failed to read client configuration file: %w", err)
		}

		cfg.Clients = append(cfg.Clients, &ccfg)
	}

	return &cfg, nil
}
