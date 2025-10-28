# Configuration Package (`config`)

Production-ready YAML configuration wrapper with Viper for Go applications. Designed for maximum developer experience with type-safe, reusable configuration management.

## Features

- **Viper-powered**: Battle-tested YAML/JSON configuration management
- **Type-safe**: Multiple typed getters for type safety
- **Environment overrides**: Automatic environment variable binding with configurable prefix
- **Multi-environment support**: Load environment-specific configs (e.g., `config.production.yaml`)
- **Global singleton**: Optional global config instance for easy access
- **Custom loaders**: Extensible architecture for custom config sources
- **Thread-safe**: Built-in RWMutex for concurrent access
- **Developer-friendly**: Comprehensive error handling and sensible defaults
- **Zero boilerplate**: Minimal setup required

## Installation

```bash
go get github.com/cubetiqlabs/gopkg
```

## Quick Start

### Basic Usage

```go
package main

import (
	"github.com/cubetiqlabs/gopkg/config"
)

func main() {
	// Create config with defaults
	// Looks for config.yaml in current directory
	cfg, err := config.New(nil)
	if err != nil {
		panic(err)
	}

	// Get values
	port := cfg.GetInt("server.port")
	host := cfg.GetString("server.host")
	debug := cfg.GetBool("debug")
}
```

### With Options

```go
cfg, err := config.New(&config.Options{
	ConfigPath: "./config",      // Directory containing config files
	ConfigName: "app",           // File name without extension
	ConfigType: "yaml",          // File type
	Env:        "production",    // Load config.production.yaml
	EnvPrefix:  "APP",           // Environment variable prefix
})
if err != nil {
	panic(err)
}
```

### Global Config

```go
// In main()
cfg, err := config.New(&config.Options{
	ConfigPath: "./config",
	Env:        os.Getenv("ENV"),
	EnvPrefix:  "APP",
})
if err != nil {
	panic(err)
}
config.SetGlobal(cfg)

// Use anywhere
func SomeFunction() {
	port := config.Global().GetInt("server.port")
}
```

## Configuration Files

### Directory Structure

```
app/
├── config/
│   ├── config.yaml           # Base config
│   ├── config.development.yaml
│   ├── config.staging.yaml
│   └── config.production.yaml
└── main.go
```

### Example config.yaml

```yaml
server:
  host: localhost
  port: 8080
  timeout: 30s

database:
  host: localhost
  port: 5432
  name: myapp

logging:
  level: info
  development: false

features:
  auth: true
  ratelimit: true
  caching: false
```

### Environment-Specific Override (config.production.yaml)

```yaml
server:
  host: 0.0.0.0
  port: 3000

logging:
  level: warn
  development: false

database:
  host: db.production.local
```

## Type-Safe Configuration

Use `Unmarshal()` or `UnmarshalKey()` for type-safe configuration:

```go
type Config struct {
	Server struct {
		Host    string        `mapstructure:"host"`
		Port    int           `mapstructure:"port"`
		Timeout time.Duration `mapstructure:"timeout"`
	} `mapstructure:"server"`
	Database struct {
		Host string `mapstructure:"host"`
		Port int    `mapstructure:"port"`
		Name string `mapstructure:"name"`
	} `mapstructure:"database"`
	Logging struct {
		Level       string `mapstructure:"level"`
		Development bool   `mapstructure:"development"`
	} `mapstructure:"logging"`
}

var appConfig Config
if err := cfg.Unmarshal(&appConfig); err != nil {
	panic(err)
}

// Use with type safety
port := appConfig.Server.Port
```

## API Reference

### Reading Values

```go
// Basic getters with type safety
cfg.GetString("key")        // Returns "" if not found
cfg.GetInt("key")           // Returns 0 if not found
cfg.GetFloat64("key")       // Returns 0.0 if not found
cfg.GetBool("key")          // Returns false if not found
cfg.GetDuration("key")      // Returns 0 if not found

// Collections
cfg.GetStringSlice("key")   // Returns []string{}
cfg.GetIntSlice("key")      // Returns []int{}
cfg.GetStringMap("key")     // Returns map[string]interface{}

// With defaults
cfg.GetStringOrDefault("key", "default")
cfg.GetIntOrDefault("key", 3000)
cfg.GetBoolOrDefault("key", false)

// Must functions (panic if not found)
cfg.MustGet("key")
cfg.MustGetString("key")
cfg.MustGetInt("key")

// Unmarshal to struct
var config ServerConfig
cfg.UnmarshalKey("server", &config)
```

### Checking Keys

```go
cfg.IsSet("key")         // Check if key exists in config file
cfg.IsSetOrEnv("key")    // Check if key exists in config or env
```

### Runtime Configuration

```go
cfg.Set("key", "value")   // Set/override at runtime
cfg.AllSettings()         // Get all settings as map
```

## Environment Variables

Environment variables automatically override config file values:

```go
cfg, _ := config.New(&config.Options{
	EnvPrefix: "APP",
})

// With EnvPrefix, these env vars are supported:
// APP_SERVER_HOST      -> server.host
// APP_SERVER_PORT      -> server.port
// APP_DATABASE_HOST    -> database.host
// APP_LOGGING_LEVEL    -> logging.level
```

Example:

```bash
APP_SERVER_PORT=9000 APP_LOGGING_LEVEL=debug ./app
```

## Custom Loaders

Extend configuration from custom sources:

```go
// Custom loader function
func loadFromDatabase(cfg *config.Config) error {
	// Fetch config from database
	dbConfig, err := db.GetConfig()
	if err != nil {
		return err
	}
	
	// Set values in config
	cfg.Set("feature_flags", dbConfig.FeatureFlags)
	return nil
}

cfg, err := config.New(&config.Options{
	Loaders: []config.Loader{loadFromDatabase},
})
```

## File Change Watching

Watch for configuration file changes:

```go
cfg, _ := config.New(nil)

cfg.WatchConfig()
cfg.Watch(func() {
	log.Println("Configuration changed!")
	// Reload services using the config
})
```

## Testing

```go
func TestMyFunction(t *testing.T) {
	cfg, _ := config.New(nil)
	
	// Override values for testing
	cfg.Set("database.host", "test-localhost")
	cfg.Set("debug", true)
	
	// Test your function
	result := MyFunction(cfg)
	assert.Equal(t, expected, result)
}
```

## Best Practices

1. **Single Config Instance**: Use global config or dependency injection
2. **Validate Early**: Unmarshal to typed structs in `main()` to catch errors early
3. **Use Type-Safe Getters**: Prefer typed getters over `Get()` where possible
4. **Separate Configs**: Use different files for different environments
5. **Document Defaults**: Comment required vs optional keys
6. **Environment Variables**: Use for sensitive data (passwords, API keys)
7. **Fail Fast**: Use `MustGet*()` for required configuration

## Example: Complete Application Setup

```go
package main

import (
	"github.com/cubetiqlabs/gopkg/config"
	"go.uber.org/zap"
)

type AppConfig struct {
	Server struct {
		Host string        `mapstructure:"host"`
		Port int           `mapstructure:"port"`
	} `mapstructure:"server"`
	Database struct {
		Host string `mapstructure:"host"`
		Port int    `mapstructure:"port"`
	} `mapstructure:"database"`
}

func main() {
	// Initialize config
	cfg, err := config.New(&config.Options{
		ConfigPath: "./config",
		Env:        getEnv(),
		EnvPrefix:  "APP",
	})
	if err != nil {
		panic(err)
	}
	
	// Validate and parse typed config
	var appConfig AppConfig
	if err := cfg.Unmarshal(&appConfig); err != nil {
		panic(err)
	}
	
	// Make it globally available
	config.SetGlobal(cfg)
	
	// Start application
	startServer(appConfig)
}

func getEnv() string {
	if env := os.Getenv("ENV"); env != "" {
		return env
	}
	return "development"
}

func startServer(cfg AppConfig) {
	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	log.Printf("Starting server on %s", addr)
	// ...
}
```

## Integration with Other Packages

```go
import (
	"github.com/cubetiqlabs/gopkg/config"
	"github.com/cubetiqlabs/gopkg/logging"
)

func main() {
	cfg, _ := config.New(nil)
	
	// Use with logging
	logLevel := cfg.GetStringOrDefault("logging.level", "info")
	logger, _ := logging.Init(logLevel, false)
	defer logger.Sync()
	
	config.SetGlobal(cfg)
	logger.Info("Config initialized")
}
```

## Troubleshooting

### Config file not found
```go
cfg, err := config.New(&config.Options{
	ConfigPath: "./config",  // Make sure this directory exists
	ConfigName: "config",
})
// Not an error - will use environment variables and defaults
```

### Environment variables not working
```go
cfg, _ := config.New(&config.Options{
	EnvPrefix: "APP",  // Required for env var binding
})

// Check env var naming:
// APP_DATABASE_HOST matches database.host (dots become underscores)
```

### Getting stale values
```go
// Reload config at runtime
cfg.WatchConfig()
cfg.Watch(func() {
	log.Println("Config reloaded")
})
```

## License

MIT - See LICENSE file for details
