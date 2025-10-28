package config

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewWithDefaults(t *testing.T) {
	cfg, err := New(nil)
	assert.NoError(t, err)
	assert.NotNil(t, cfg)
}

func TestGetString(t *testing.T) {
	cfg, err := New(nil)
	require.NoError(t, err)
	cfg.Set("app.name", "test-app")
	assert.Equal(t, "test-app", cfg.GetString("app.name"))
}

func TestGetInt(t *testing.T) {
	cfg, err := New(nil)
	require.NoError(t, err)
	cfg.Set("server.port", 8080)
	assert.Equal(t, 8080, cfg.GetInt("server.port"))
}

func TestGetBool(t *testing.T) {
	cfg, err := New(nil)
	require.NoError(t, err)
	cfg.Set("debug", true)
	assert.True(t, cfg.GetBool("debug"))
}

func TestGetDuration(t *testing.T) {
	cfg, err := New(nil)
	require.NoError(t, err)
	cfg.Set("timeout", "5s")
	assert.Equal(t, 5*time.Second, cfg.GetDuration("timeout"))
}

func TestUnmarshal(t *testing.T) {
	cfg, err := New(nil)
	require.NoError(t, err)
	cfg.Set("server.host", "localhost")
	cfg.Set("server.port", 8080)

	type ServerConfig struct {
		Host string `mapstructure:"host"`
		Port int    `mapstructure:"port"`
	}

	var result ServerConfig
	err = cfg.UnmarshalKey("server", &result)
	require.NoError(t, err)
	assert.Equal(t, "localhost", result.Host)
	assert.Equal(t, 8080, result.Port)
}

func TestGetOrDefault(t *testing.T) {
	cfg, err := New(nil)
	require.NoError(t, err)
	assert.Equal(t, "default", cfg.GetStringOrDefault("nonexistent", "default"))
	assert.Equal(t, 3000, cfg.GetIntOrDefault("nonexistent", 3000))
}

func TestEnvironmentVariables(t *testing.T) {
	os.Setenv("APP_DATABASE_HOST", "env-localhost")
	defer os.Unsetenv("APP_DATABASE_HOST")

	cfg, err := New(&Options{
		EnvPrefix: "APP",
	})
	require.NoError(t, err)
	assert.Equal(t, "env-localhost", cfg.GetString("database.host"))
}

func TestCustomLoader(t *testing.T) {
	loader := func(cfg *Config) error {
		cfg.Set("loaded", true)
		return nil
	}

	cfg, err := New(&Options{
		Loaders: []Loader{loader},
	})
	require.NoError(t, err)
	assert.True(t, cfg.GetBool("loaded"))
}

func TestGlobalConfig(t *testing.T) {
	globalConfig = nil
	cfg, err := New(&Options{})
	require.NoError(t, err)
	SetGlobal(cfg)
	assert.Equal(t, cfg, Global())
}
