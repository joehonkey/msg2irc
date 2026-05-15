package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type TLSConfig struct {
	Mode       string `yaml:"mode"`        // auto, manual, off
	Domain     string `yaml:"domain"`      // required for auto
	CertFile   string `yaml:"cert"`        // required for manual
	KeyFile    string `yaml:"key"`         // required for manual
	ListenAddr string `yaml:"listen_addr"` // default :443 for auto/manual
	CacheDir   string `yaml:"cache_dir"`   // cert cache for auto, default ./certs
}

type Config struct {
	ListenAddr string    `yaml:"listen_addr"` // used when tls.mode is off
	Token      string    `yaml:"token"`
	TLS        TLSConfig `yaml:"tls"`
	IRC        struct {
		Server      string   `yaml:"server"`
		Port        int      `yaml:"port"`
		Nick        string   `yaml:"nick"`
		TLS         bool     `yaml:"tls"`
		TLSHostname string   `yaml:"tls_hostname"`
		Channels    []string `yaml:"channels"`
	} `yaml:"irc"`
}

func Load(path string) (*Config, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var cfg Config
	if err := yaml.NewDecoder(f).Decode(&cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
