package watchmen

import (
	"os"
	"strings"
)

// Config ...
type Config struct {
	MainnetEnabled bool

	Port string
}

// NewConfig ...
func NewConfig() Config {
	cfg := Config{}

	cfg.MainnetEnabled = false
	if p := strings.ToLower(os.Getenv("MAINNET_ENABLED")); p == "true" {
		cfg.MainnetEnabled = true
	}

	cfg.Port = "8080"
	if p := os.Getenv("PORT"); p != "" {
		cfg.Port = p
	}

	return cfg
}
