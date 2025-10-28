package config

import (
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
)

// Config wraps Viper for type-safe configuration management.
// It provides a clean, reusable interface for loading and accessing configuration.
type Config struct {
	viper *viper.Viper
	mu    sync.RWMutex
}

// Loader is a function that loads configuration from an external source.
// Useful for custom configuration sources (databases, remote services, etc.)
type Loader func(cfg *Config) error

// Options configures Config initialization behavior.
type Options struct {
	// ConfigPath is the directory containing config files (default: ".")
	ConfigPath string
	// ConfigName is the config file name without extension (default: "config")
	ConfigName string
	// ConfigType is the file type (yaml, json, toml, etc.) (default: "yaml")
	ConfigType string
	// Env specifies the environment name for loading env-specific configs (default: "")
	// If set, loads config.{Env}.yaml after config.yaml
	Env string
	// EnvPrefix specifies the prefix for environment variables (default: "")
	// All environment variables will be auto-bound with this prefix
	EnvPrefix string
	// AutoEnvEnabled enables automatic binding of all environment variables (default: true)
	AutoEnvEnabled bool
	// LookupsEnv enables case-insensitive environment variable lookup (default: true)
	LookupsEnv bool
	// Loaders are custom configuration loaders to execute after initial load (default: nil)
	Loaders []Loader
}

var (
	// Global config instance
	globalConfig *Config
	globalMu     sync.Once
)

// New creates a new Config instance with default options.
// Default options:
//   - ConfigPath: "."
//   - ConfigName: "config"
//   - ConfigType: "yaml"
//   - EnvPrefix: ""
//   - AutoEnvEnabled: true
//   - LookupsEnv: true
//
// Example:
//
//	cfg, err := config.New(&config.Options{
//	    ConfigPath: "./config",
//	    Env: "production",
//	    EnvPrefix: "APP",
//	})
//	if err != nil {
//	    panic(err)
//	}
func New(opts *Options) (*Config, error) {
	if opts == nil {
		opts = &Options{}
	}

	// Set defaults
	if opts.ConfigPath == "" {
		opts.ConfigPath = "."
	}
	if opts.ConfigName == "" {
		opts.ConfigName = "config"
	}
	if opts.ConfigType == "" {
		opts.ConfigType = "yaml"
	}
	opts.AutoEnvEnabled = true // enabled by default
	opts.LookupsEnv = true     // enabled by default

	v := viper.New()

	// Configure paths
	v.AddConfigPath(opts.ConfigPath)
	v.SetConfigName(opts.ConfigName)
	v.SetConfigType(opts.ConfigType)

	// Configure environment variables
	if opts.EnvPrefix != "" {
		v.SetEnvPrefix(opts.EnvPrefix)
	}
	if opts.AutoEnvEnabled {
		v.AutomaticEnv()
	}
	if opts.LookupsEnv {
		v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	}

	cfg := &Config{viper: v}

	// Load base config
	if err := cfg.loadConfig(); err != nil {
		return nil, err
	}

	// Load environment-specific config if specified
	if opts.Env != "" {
		if err := cfg.loadEnvConfig(opts.Env); err != nil {
			return nil, err
		}
	}

	// Execute custom loaders
	for _, loader := range opts.Loaders {
		if err := loader(cfg); err != nil {
			return nil, fmt.Errorf("config loader failed: %w", err)
		}
	}

	return cfg, nil
}

// Global returns the global Config instance. Panics if not initialized.
// Use SetGlobal() to initialize the global instance.
//
// Example:
//
//	cfg := config.Global()
func Global() *Config {
	if globalConfig == nil {
		panic("global config not initialized, call config.SetGlobal() first")
	}
	return globalConfig
}

// SetGlobal sets the global Config instance. Safe to call multiple times; first call wins.
// Useful for initialization in main().
//
// Example:
//
//	cfg, err := config.New(&config.Options{Env: "production"})
//	if err != nil {
//	    panic(err)
//	}
//	config.SetGlobal(cfg)
func SetGlobal(cfg *Config) {
	globalMu.Do(func() {
		globalConfig = cfg
	})
}

// loadConfig loads the base configuration file.
func (c *Config) loadConfig() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if err := c.viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			return nil
		}
		return fmt.Errorf("failed to read config: %w", err)
	}

	return nil
}

// loadEnvConfig loads environment-specific configuration.
// It looks for files like config.production.yaml
func (c *Config) loadEnvConfig(env string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	envConfigName := fmt.Sprintf("%s.%s", c.viper.ConfigFileUsed(), env)
	c.viper.SetConfigName(envConfigName)

	if err := c.viper.MergeInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			return nil
		}
		return fmt.Errorf("failed to read env config: %w", err)
	}

	return nil
}

// Get returns a configuration value as interface{}
func (c *Config) Get(key string) interface{} {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.viper.Get(key)
}

// GetString returns a configuration value as string
func (c *Config) GetString(key string) string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.viper.GetString(key)
}

// GetInt returns a configuration value as int
func (c *Config) GetInt(key string) int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.viper.GetInt(key)
}

// GetFloat64 returns a configuration value as float64
func (c *Config) GetFloat64(key string) float64 {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.viper.GetFloat64(key)
}

// GetBool returns a configuration value as bool
func (c *Config) GetBool(key string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.viper.GetBool(key)
}

// GetDuration returns a configuration value as time.Duration
func (c *Config) GetDuration(key string) time.Duration {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.viper.GetDuration(key)
}

// GetStringSlice returns a configuration value as []string
func (c *Config) GetStringSlice(key string) []string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.viper.GetStringSlice(key)
}

// GetIntSlice returns a configuration value as []int
func (c *Config) GetIntSlice(key string) []int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.viper.GetIntSlice(key)
}

// GetStringMap returns a configuration value as map[string]interface{}
func (c *Config) GetStringMap(key string) map[string]interface{} {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.viper.GetStringMap(key)
}

// GetStringMapString returns a configuration value as map[string]string
func (c *Config) GetStringMapString(key string) map[string]string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.viper.GetStringMapString(key)
}

// GetStringMapStringSlice returns a configuration value as map[string][]string
func (c *Config) GetStringMapStringSlice(key string) map[string][]string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.viper.GetStringMapStringSlice(key)
}

// Unmarshal unmarshals configuration into a struct.
// Use this for type-safe configuration handling.
func (c *Config) Unmarshal(rawVal interface{}) error {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.viper.Unmarshal(rawVal)
}

// UnmarshalKey unmarshals a configuration key into a struct.
func (c *Config) UnmarshalKey(key string, rawVal interface{}) error {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.viper.UnmarshalKey(key, rawVal)
}

// IsSet returns whether a key is set in configuration.
func (c *Config) IsSet(key string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.viper.IsSet(key)
}

// IsSetOrEnv returns whether a key is set in configuration or as environment variable.
func (c *Config) IsSetOrEnv(key string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.viper.IsSet(key) {
		return true
	}

	envKey := strings.ToUpper(strings.ReplaceAll(key, ".", "_"))
	_, exists := os.LookupEnv(envKey)
	return exists
}

// AllSettings returns all configuration settings.
func (c *Config) AllSettings() map[string]interface{} {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.viper.AllSettings()
}

// Set sets a configuration value at runtime.
func (c *Config) Set(key string, value interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.viper.Set(key, value)
}

// Watch registers a callback to be called when configuration changes.
func (c *Config) Watch(callback func()) {
	c.viper.OnConfigChange(func(in fsnotify.Event) {
		callback()
	})
}

// WatchConfig enables watching for configuration file changes.
func (c *Config) WatchConfig() {
	c.viper.WatchConfig()
}

// Viper returns the underlying Viper instance for advanced operations.
func (c *Config) Viper() *viper.Viper {
	return c.viper
}

// MustGet is like Get but panics if the key is not found.
func (c *Config) MustGet(key string) interface{} {
	if !c.IsSet(key) {
		panic(fmt.Sprintf("required config key not found: %s", key))
	}
	return c.Get(key)
}

// MustGetString is like GetString but panics if the key is not found.
func (c *Config) MustGetString(key string) string {
	if !c.IsSet(key) {
		panic(fmt.Sprintf("required config key not found: %s", key))
	}
	return c.GetString(key)
}

// MustGetInt is like GetInt but panics if the key is not found.
func (c *Config) MustGetInt(key string) int {
	if !c.IsSet(key) {
		panic(fmt.Sprintf("required config key not found: %s", key))
	}
	return c.GetInt(key)
}

// GetOrDefault returns a value or a default if not found.
func (c *Config) GetOrDefault(key string, defaultVal interface{}) interface{} {
	if c.IsSet(key) {
		return c.Get(key)
	}
	return defaultVal
}

// GetStringOrDefault returns a string value or a default if not found.
func (c *Config) GetStringOrDefault(key string, defaultVal string) string {
	if c.IsSet(key) {
		return c.GetString(key)
	}
	return defaultVal
}

// GetIntOrDefault returns an int value or a default if not found.
func (c *Config) GetIntOrDefault(key string, defaultVal int) int {
	if c.IsSet(key) {
		return c.GetInt(key)
	}
	return defaultVal
}

// GetBoolOrDefault returns a bool value or a default if not found.
func (c *Config) GetBoolOrDefault(key string, defaultVal bool) bool {
	if c.IsSet(key) {
		return c.GetBool(key)
	}
	return defaultVal
}
