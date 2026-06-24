package config

import (
	"embed"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/fair-n-square-co/core/internal/core/db"
	"github.com/fair-n-square-co/core/internal/core/logger"
	"github.com/spf13/viper"
)

//go:embed *.yml
var configFiles embed.FS

// Config is the fully-resolved application configuration. Sub-configs are
// composed from the packages that own them (e.g. db.DBConfig) so each module
// defines its own settings and they are populated here in one place.
type Config struct {
	Port   int
	Logger logger.LogConfig
	Db     db.DBConfig

	viperReader *viper.Viper
}

// LoadConfig resolves config in two passes into the same struct: embedded YAML
// first, then environment variables (prefixed CORE_) which override the YAML.
func LoadConfig() (*Config, error) {
	config := &Config{
		Port: 8080,
		Logger: logger.LogConfig{
			Level:  slog.LevelInfo,
			Format: "json",
		},
	}

	if err := config.readViperConfig("yaml"); err != nil {
		return nil, fmt.Errorf("config error: failed to read yaml config: %w", err)
	}
	if err := config.readViperConfig("env"); err != nil {
		return nil, fmt.Errorf("config error: failed to read env config: %w", err)
	}

	return config, nil
}

// readViperConfig runs a single resolution pass and unmarshals into the shared
// Config. ExperimentalBindStruct lets AutomaticEnv bind to struct fields without
// explicit BindEnv calls; KeyDelimiter("_") makes nested keys map to env vars
// like CORE_DB_CONNECTIONSTRING.
func (c *Config) readViperConfig(configType string) error {
	c.viperReader = viper.NewWithOptions(
		viper.ExperimentalBindStruct(),
		viper.KeyDelimiter("_"),
	)

	if configType == "env" {
		c.viperReader.SetConfigType("env")
		c.viperReader.SetEnvPrefix("core")
		c.viperReader.AutomaticEnv()
	} else {
		configFile, err := getConfigFile()
		if err != nil {
			return fmt.Errorf("failed to get config file: %w", err)
		}
		c.viperReader.SetConfigType("yaml")
		if err := c.viperReader.ReadConfig(strings.NewReader(configFile)); err != nil {
			if _, ok := err.(viper.ConfigFileNotFoundError); ok {
				slog.Info("config file not found, using defaults", slog.String("configType", configType))
				return nil
			}
			return fmt.Errorf("failed to read config: %w", err)
		}
	}

	if err := c.viperReader.Unmarshal(c); err != nil {
		return fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return nil
}

// getConfigFile returns the embedded YAML for the current ENV ("production"
// selects config.prod.yml; everything else uses config.yml).
func getConfigFile() (string, error) {
	name := "config.yml"
	if os.Getenv("CORE_ENV") == "production" {
		name = "config.prod.yml"
	}
	b, err := configFiles.ReadFile(name)
	if err != nil {
		return "", fmt.Errorf("read embedded %s: %w", name, err)
	}
	return string(b), nil
}
