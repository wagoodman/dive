package config

type loggingConfig struct {
	Enabled bool   `mapstructure:"enabled"`
	Level   string `mapstructure:"level"`
	Path    string `mapstructure:"path"`
}
