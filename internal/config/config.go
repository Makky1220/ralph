package config

import (
	"errors"
	"io/fs"
	"os"

	toml "github.com/pelletier/go-toml/v2"
)

// Config represents the ralph.toml project configuration.
type Config struct {
	Doctor DoctorConfig `toml:"doctor"`
}

// DoctorConfig holds doctor check settings.
type DoctorConfig struct {
	RequireClaudeCLI bool `toml:"require_claude_cli"`
	RequireGo        bool `toml:"require_go"`
}

// Default returns a Config with sensible defaults.
func Default() Config {
	return Config{
		Doctor: DoctorConfig{
			RequireClaudeCLI: true,
			RequireGo:        false,
		},
	}
}

// Load reads ralph.toml from path, falling back to defaults when absent.
func Load(path string) (Config, error) {
	cfg := Default()

	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return cfg, nil
		}
		return cfg, err
	}

	if err := toml.Unmarshal(data, &cfg); err != nil {
		return cfg, err
	}

	return cfg, nil
}
