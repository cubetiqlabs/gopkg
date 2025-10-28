package config

// Example demonstrates best practices for using the config package.
// This example shows how to structure your application with type-safe configuration.

import (
	"context"
	"fmt"
)

// AppConfig represents the entire application configuration
type AppConfig struct {
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
	Logging  LoggingConfig  `mapstructure:"logging"`
	Features FeatureFlags   `mapstructure:"features"`
}

// ServerConfig holds server-related configuration
type ServerConfig struct {
	Host string `mapstructure:"host"`
	Port int    `mapstructure:"port"`
}

// DatabaseConfig holds database-related configuration
type DatabaseConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Name     string `mapstructure:"name"`
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
}

// LoggingConfig holds logging-related configuration
type LoggingConfig struct {
	Level       string `mapstructure:"level"`
	Development bool   `mapstructure:"development"`
}

// FeatureFlags holds feature flag configuration
type FeatureFlags struct {
	AuthEnabled      bool `mapstructure:"auth_enabled"`
	RateLimitEnabled bool `mapstructure:"ratelimit_enabled"`
}

// Example_BasicUsage shows basic usage of the config package
func Example_BasicUsage() {
	// Create a config with default options
	cfg, err := New(nil)
	if err != nil {
		panic(err)
	}

	// Get individual values
	cfg.Set("server.host", "localhost")
	cfg.Set("server.port", 8080)

	host := cfg.GetString("server.host")
	port := cfg.GetInt("server.port")

	fmt.Printf("Server: %s:%d\n", host, port)
}

// Example_TypeSafeConfig shows how to use type-safe configuration
func Example_TypeSafeConfig() {
	cfg, err := New(nil)
	if err != nil {
		panic(err)
	}

	// Set test data
	cfg.Set("server.host", "localhost")
	cfg.Set("server.port", 8080)
	cfg.Set("database.host", "localhost")
	cfg.Set("database.port", 5432)
	cfg.Set("database.name", "myapp")
	cfg.Set("logging.level", "info")
	cfg.Set("logging.development", false)
	cfg.Set("features.auth_enabled", true)
	cfg.Set("features.ratelimit_enabled", false)

	// Unmarshal entire config
	var appConfig AppConfig
	if err := cfg.Unmarshal(&appConfig); err != nil {
		panic(err)
	}

	fmt.Printf("Config: %+v\n", appConfig)
}

// Example_EnvironmentOverrides shows how environment variables override config
func Example_EnvironmentOverrides() {
	// With EnvPrefix "APP", environment variables like:
	// APP_SERVER_PORT=9000
	// APP_DATABASE_HOST=prod-db.local
	// will override the config file values

	cfg, err := New(&Options{
		ConfigPath: ".",
		ConfigName: "config",
		ConfigType: "yaml",
		EnvPrefix:  "APP",
	})
	if err != nil {
		panic(err)
	}

	// Env var APP_SERVER_PORT=9000 would override this
	cfg.Set("server.port", 8080)
	port := cfg.GetInt("server.port")
	fmt.Printf("Server port: %d\n", port)
}

// Example_WithGlobalConfig shows using a global config instance
func Example_WithGlobalConfig() {
	cfg, err := New(nil)
	if err != nil {
		panic(err)
	}

	// Set as global
	SetGlobal(cfg)

	// Can be accessed from anywhere
	globalCfg := Global()
	globalCfg.Set("app.name", "MyApp")
	name := globalCfg.GetString("app.name")
	fmt.Printf("App name: %s\n", name)
}

// Example_CustomLoader shows how to use custom loaders
func Example_CustomLoader() {
	// Custom loader that could load from database, remote service, etc.
	customLoader := func(cfg *Config) error {
		// Simulate loading feature flags from external service
		featureFlags := map[string]bool{
			"new_ui":      true,
			"beta_api":    false,
			"maintenance": false,
		}

		cfg.Set("runtime.feature_flags", featureFlags)
		return nil
	}

	cfg, err := New(&Options{
		Loaders: []Loader{customLoader},
	})
	if err != nil {
		panic(err)
	}

	flags := cfg.GetStringMap("runtime.feature_flags")
	fmt.Printf("Feature flags: %+v\n", flags)
}

// Example_DefaultValues shows using GetOrDefault for optional config
func Example_DefaultValues() {
	cfg, err := New(nil)
	if err != nil {
		panic(err)
	}

	// Using GetOrDefault for optional configuration
	timeout := cfg.GetIntOrDefault("server.timeout", 30)
	maxConnections := cfg.GetIntOrDefault("server.max_connections", 100)
	debug := cfg.GetBoolOrDefault("debug", false)

	fmt.Printf("Timeout: %d, MaxConnections: %d, Debug: %v\n", timeout, maxConnections, debug)
}

// Example_DependencyInjection shows proper dependency injection pattern
func Example_DependencyInjection() {
	cfg, err := New(nil)
	if err != nil {
		panic(err)
	}

	// Unmarshal config in main()
	var appConfig AppConfig
	if err := cfg.Unmarshal(&appConfig); err != nil {
		panic(err)
	}

	// Pass config to functions that need it
	setupDatabase(appConfig.Database)
	setupServer(appConfig.Server)
}

func setupDatabase(cfg DatabaseConfig) {
	fmt.Printf("Setting up database: %s:%d\n", cfg.Host, cfg.Port)
}

func setupServer(cfg ServerConfig) {
	fmt.Printf("Starting server on %s:%d\n", cfg.Host, cfg.Port)
}

// Example_WithContext shows using config with context
func Example_WithContext(ctx context.Context) {
	cfg, err := New(nil)
	if err != nil {
		panic(err)
	}

	// Store config in context for request handlers
	ctx = context.WithValue(ctx, "config", cfg)

	// Later, retrieve it
	retrievedCfg := ctx.Value("config").(*Config)
	port := retrievedCfg.GetInt("server.port")
	fmt.Printf("Port from context: %d\n", port)
}
