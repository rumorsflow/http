package http

import (
	"crypto/tls"
	"golang.org/x/crypto/acme"
	"golang.org/x/crypto/acme/autocert"
	"net/http"
	"time"
)

type Config struct {
	DirCache          string        `mapstructure:"dir_cache"`
	HostWhitelist     []string      `mapstructure:"host_whitelist"`
	CertFile          string        `mapstructure:"cert_file"`
	KeyFile           string        `mapstructure:"key_file"`
	Address           string        `mapstructure:"address"`
	Middleware        []string      `mapstructure:"middleware"`
	MaxHeaderBytes    int           `mapstructure:"max_header_bytes"`
	ReadHeaderTimeout time.Duration `mapstructure:"read_header_timeout"`
	ReadTimeout       time.Duration `mapstructure:"read_timeout"`
	WriteTimeout      time.Duration `mapstructure:"write_timeout"`
	IdleTimeout       time.Duration `mapstructure:"idle_timeout"`
	Static            []struct {
		Dir     string   `mapstructure:"dir"`
		Pattern string   `mapstructure:"pattern"`
		Methods []string `mapstructure:"methods"`
	} `mapstructure:"static"`
}

func (cfg *Config) InitDefault() {
	if cfg.Address == "" {
		cfg.Address = ":0"
	}
}

func (cfg *Config) IsTLS() bool {
	return (cfg.DirCache != "" && len(cfg.HostWhitelist) > 0) || (cfg.CertFile != "" && cfg.KeyFile != "")
}

func (cfg *Config) TLSConfig() *tls.Config {
	if cfg.IsTLS() {
		if cfg.DirCache == "" || len(cfg.HostWhitelist) == 0 {
			return new(tls.Config)
		} else {
			autoTLSManager := autocert.Manager{
				Prompt:     autocert.AcceptTOS,
				HostPolicy: autocert.HostWhitelist(cfg.HostWhitelist...),
				Cache:      autocert.DirCache(cfg.DirCache),
			}

			return &tls.Config{
				GetCertificate: autoTLSManager.GetCertificate,
				NextProtos:     []string{acme.ALPNProto},
			}
		}
	}
	return nil
}

func (cfg *Config) Server(handler http.Handler) *http.Server {
	return &http.Server{
		Handler:           handler,
		TLSConfig:         cfg.TLSConfig(),
		Addr:              cfg.Address,
		MaxHeaderBytes:    cfg.MaxHeaderBytes,
		ReadHeaderTimeout: cfg.ReadHeaderTimeout,
		ReadTimeout:       cfg.ReadTimeout,
		WriteTimeout:      cfg.WriteTimeout,
		IdleTimeout:       cfg.IdleTimeout,
	}
}
