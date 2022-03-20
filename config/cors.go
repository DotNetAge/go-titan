package config

import (
	"github.com/gorilla/handlers"
	"net/http"
)

type CORSConfig struct {
	Origins []string `mapstructure:"origins"`
	Methods []string `mapstructure:"methods"`
	Headers []string `mapstructure:"headers"`
}


func (cfg *CORSConfig) Allows(h http.Handler) http.Handler {

	opts := make([]handlers.CORSOption, 0)
	if cfg.Headers != nil {
		opts = append(opts, handlers.AllowedHeaders(cfg.Headers))
	}

	if cfg.Methods != nil {
		opts = append(opts, handlers.AllowedMethods(cfg.Methods))
	}

	if cfg.Origins != nil {
		opts = append(opts, handlers.AllowedOrigins(cfg.Origins))
	}

	if len(opts) > 0 {
		return handlers.CORS(opts...)(h)
	}

	return h
}
