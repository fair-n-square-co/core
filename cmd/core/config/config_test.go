package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestLoadConfig_Defaults verifies the embedded YAML defaults are applied when
// no environment overrides are present.
func TestLoadConfig_Defaults(t *testing.T) {
	t.Setenv("CORE_ENV", "")
	cfg, err := LoadConfig()
	require.NoError(t, err)

	assert.Equal(t, 8080, cfg.Port)
	assert.Equal(t, "json", cfg.Logger.Format)
	assert.Empty(t, cfg.Db.ConnString)
}

// TestLoadConfig_EnvOverrides verifies the env pass overrides YAML values,
// including the composed db.DBConfig field via the CORE_ prefix.
func TestLoadConfig_EnvOverrides(t *testing.T) {
	t.Setenv("CORE_ENV", "")
	t.Setenv("CORE_PORT", "9090")
	t.Setenv("CORE_LOGGER_FORMAT", "text")
	t.Setenv("CORE_DB_CONNSTRING", "postgres://user:pass@host:5432/core?sslmode=disable")

	cfg, err := LoadConfig()
	require.NoError(t, err)

	assert.Equal(t, 9090, cfg.Port)
	assert.Equal(t, "text", cfg.Logger.Format)
	assert.Equal(t, "postgres://user:pass@host:5432/core?sslmode=disable", cfg.Db.ConnString)
}
