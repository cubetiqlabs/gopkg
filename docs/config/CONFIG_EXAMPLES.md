# Example Configuration Files

This directory contains example configuration files for the gopkg config package.

## Directory Structure

```
project/
├── config/
│   ├── config.yaml              # Base configuration (committed to git)
│   ├── config.development.yaml  # Development overrides
│   ├── config.staging.yaml      # Staging overrides
│   └── config.production.yaml   # Production overrides
└── main.go
```

## config.yaml (Base Configuration)

```yaml
server:
  host: localhost
  port: 8080
  timeout: 30s
  max_connections: 100

database:
  host: localhost
  port: 5432
  name: myapp_dev
  username: app_user
  password: dev_password
  pool_size: 10
  ssl_mode: disable

redis:
  host: localhost
  port: 6379
  database: 0
  ttl: 3600

logging:
  level: info
  development: false
  format: json

features:
  auth_enabled: true
  ratelimit_enabled: true
  caching_enabled: false
  maintenance_mode: false

email:
  smtp_host: smtp.example.com
  smtp_port: 587
  from_address: noreply@example.com
  enabled: true

cors:
  allowed_origins:
    - http://localhost:3000
  allowed_methods:
    - GET
    - POST
    - PUT
    - DELETE
  allowed_headers:
    - Content-Type
    - Authorization

jwt:
  secret: your-secret-key-change-in-production
  expiration: 3600
```

## config.development.yaml (Development Overrides)

```yaml
server:
  port: 3000
  timeout: 60s

database:
  name: myapp_dev
  password: dev_password
  pool_size: 5

logging:
  level: debug
  development: true

features:
  caching_enabled: false
  maintenance_mode: false

cors:
  allowed_origins:
    - http://localhost:3000
    - http://localhost:3001

jwt:
  secret: dev-secret-key
  expiration: 86400
```

## config.production.yaml (Production Overrides)

```yaml
server:
  host: 0.0.0.0
  port: 8080
  timeout: 30s
  max_connections: 500

database:
  host: prod-db.example.com
  port: 5432
  name: myapp_prod
  username: app_prod_user
  pool_size: 50
  ssl_mode: require

redis:
  host: prod-redis.example.com
  port: 6379
  database: 0
  ttl: 7200

logging:
  level: warn
  development: false

features:
  auth_enabled: true
  ratelimit_enabled: true
  caching_enabled: true
  maintenance_mode: false

email:
  smtp_host: smtp.sendgrid.net
  smtp_port: 587
  from_address: noreply@myapp.com
  enabled: true

cors:
  allowed_origins:
    - https://myapp.com
    - https://www.myapp.com

jwt:
  # Should be set via environment variable: APP_JWT_SECRET
  expiration: 3600
```

## Usage in Go

### main.go

```go
package main

import (
	"fmt"
	"os"

	"github.com/cubetiqlabs/gopkg/config"
)

type AppConfig struct {
	Server struct {
		Host             string `mapstructure:"host"`
		Port             int    `mapstructure:"port"`
		Timeout          string `mapstructure:"timeout"`
		MaxConnections   int    `mapstructure:"max_connections"`
	} `mapstructure:"server"`
	Database struct {
		Host     string `mapstructure:"host"`
		Port     int    `mapstructure:"port"`
		Name     string `mapstructure:"name"`
		Username string `mapstructure:"username"`
		Password string `mapstructure:"password"`
		PoolSize int    `mapstructure:"pool_size"`
		SSLMode  string `mapstructure:"ssl_mode"`
	} `mapstructure:"database"`
	Logging struct {
		Level       string `mapstructure:"level"`
		Development bool   `mapstructure:"development"`
	} `mapstructure:"logging"`
	Features struct {
		AuthEnabled      bool `mapstructure:"auth_enabled"`
		RateLimitEnabled bool `mapstructure:"ratelimit_enabled"`
		CachingEnabled   bool `mapstructure:"caching_enabled"`
		MaintenanceMode  bool `mapstructure:"maintenance_mode"`
	} `mapstructure:"features"`
	JWT struct {
		Secret     string `mapstructure:"secret"`
		Expiration int    `mapstructure:"expiration"`
	} `mapstructure:"jwt"`
}

func main() {
	// Get environment (defaults to "development")
	env := os.Getenv("ENV")
	if env == "" {
		env = "development"
	}

	// Initialize configuration
	cfg, err := config.New(&config.Options{
		ConfigPath: "./config",
		ConfigName: "config",
		ConfigType: "yaml",
		Env:        env,
		EnvPrefix:  "APP",
	})
	if err != nil {
		panic(err)
	}

	// Unmarshal to typed struct
	var appConfig AppConfig
	if err := cfg.Unmarshal(&appConfig); err != nil {
		panic(err)
	}

	// Make globally available
	config.SetGlobal(cfg)

	// Use configuration
	fmt.Printf("Starting server on %s:%d\n", appConfig.Server.Host, appConfig.Server.Port)
	fmt.Printf("Database: %s@%s:%d/%s\n",
		appConfig.Database.Username,
		appConfig.Database.Host,
		appConfig.Database.Port,
		appConfig.Database.Name,
	)
	fmt.Printf("Logging level: %s\n", appConfig.Logging.Level)
	fmt.Printf("Features - Auth: %v, RateLimit: %v, Caching: %v\n",
		appConfig.Features.AuthEnabled,
		appConfig.Features.RateLimitEnabled,
		appConfig.Features.CachingEnabled,
	)
}
```

## Environment Variable Usage

You can override any configuration value using environment variables:

```bash
# Development
ENV=development go run main.go

# Production with environment variable overrides
ENV=production \
APP_DATABASE_PASSWORD=secret123 \
APP_JWT_SECRET=production-jwt-secret \
APP_SERVER_PORT=8080 \
go run main.go
```

## Best Practices

1. **Commit base config**: Always commit `config.yaml` to version control
2. **Ignore env-specific files**: Add to `.gitignore`:
   ```
   config.*.local.yaml
   .env
   ```
3. **Use environment variables for secrets**:
   - Database passwords
   - API keys
   - JWT secrets
   - Any sensitive data

4. **Document configuration**:
   ```go
   // Example configuration in code comments
   type ServerConfig struct {
       Host string `mapstructure:"host"` // Server hostname
       Port int    `mapstructure:"port"` // Server port (default: 8080)
   }
   ```

5. **Load early, fail fast**:
   ```go
   // In main(), before using config
   var appConfig AppConfig
   if err := config.Global().Unmarshal(&appConfig); err != nil {
       panic(fmt.Sprintf("Invalid config: %v", err))
   }
   ```

6. **Use different configs per environment**:
   - Base: `config.yaml` (defaults, shared settings)
   - Environment: `config.{env}.yaml` (environment-specific overrides)
   - Runtime: Environment variables (secrets, dynamic values)
