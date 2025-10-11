# contextx

Type-safe context value management for Go applications.

## Features

- **Type-safe** context value storage and retrieval
- **Multi-tenant** application support
- **Application scoping** within tenants
- **Audit trail** support with API key tracking
- **Zero dependencies** (only standard library)

## Installation

```bash
go get github.com/cubetiqlabs/gopkg/contextx
```

## Usage

### Basic Tenant Context

```go
package main

import (
    "context"
    "fmt"
    "github.com/cubetiqlabs/gopkg/contextx"
)

func main() {
    ctx := context.Background()
    
    // Store tenant ID
    ctx = contextx.WithTenant(ctx, "tenant-123")
    
    // Retrieve tenant ID
    tenantID, ok := contextx.TenantID(ctx)
    if ok {
        fmt.Println("Tenant ID:", tenantID)
    }
}
```

### Application Context

```go
// Store application ID within tenant
ctx = contextx.WithApplication(ctx, "app-456")

// Retrieve application ID
appID, ok := contextx.AppID(ctx)
```

### Combined Auth Values

```go
import "time"

// Store all auth values together
now := time.Now()
values := contextx.TenantAuthValues{
    TenantID:  "tenant-123",
    AppID:     "app-456",
    Prefix:    "sk_live_",
    LastUsed:  &now,
    CreatedAt: &now,
}

ctx = contextx.WithTenantAuthValues(ctx, values)

// Retrieve all values
auth, ok := contextx.TenantAuth(ctx)
if ok {
    fmt.Println("Tenant:", auth.TenantID)
    fmt.Println("App:", auth.AppID)
    fmt.Println("API Key Prefix:", auth.Prefix)
}
```

### API Key Actor Tracking

```go
// Store API key prefix for audit trails
ctx = contextx.WithAPIKeyPrefix(ctx, "sk_test_")

// Retrieve actor information
actor, ok := contextx.APIKeyActor(ctx)
```

## Use Cases

### Multi-Tenant Web Applications

```go
func authMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        apiKey := r.Header.Get("X-API-Key")
        
        // Validate API key and extract tenant
        tenant, err := validateAPIKey(apiKey)
        if err != nil {
            http.Error(w, "Unauthorized", 401)
            return
        }
        
        // Inject tenant into context
        ctx := contextx.WithTenant(r.Context(), tenant.ID)
        
        // Continue with modified context
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}

func handleRequest(w http.ResponseWriter, r *http.Request) {
    // Extract tenant from context
    tenantID, ok := contextx.TenantID(r.Context())
    if !ok {
        http.Error(w, "No tenant context", 500)
        return
    }
    
    // Use tenant ID for data isolation
    data := fetchTenantData(tenantID)
    json.NewEncoder(w).Encode(data)
}
```

### Fiber Framework Integration

```go
import (
    "github.com/gofiber/fiber/v2"
    "github.com/cubetiqlabs/gopkg/contextx"
)

func authMiddleware(c *fiber.Ctx) error {
    apiKey := c.Get("X-API-Key")
    
    // Validate and extract tenant
    tenant, err := validateAPIKey(apiKey)
    if err != nil {
        return fiber.ErrUnauthorized
    }
    
    // Inject into Fiber context
    ctx := contextx.WithTenant(c.UserContext(), tenant.ID)
    c.SetUserContext(ctx)
    
    return c.Next()
}

func getData(c *fiber.Ctx) error {
    tenantID, ok := contextx.TenantID(c.UserContext())
    if !ok {
        return fiber.ErrUnauthorized
    }
    
    return c.JSON(fiber.Map{
        "tenant": tenantID,
        "data":   fetchData(tenantID),
    })
}
```

### Service Layer

```go
type UserService struct {
    db *sql.DB
}

func (s *UserService) GetUser(ctx context.Context, userID string) (*User, error) {
    // Extract tenant for data isolation
    tenantID, ok := contextx.TenantID(ctx)
    if !ok {
        return nil, errors.New("no tenant context")
    }
    
    // Query with tenant isolation
    query := "SELECT * FROM users WHERE id = $1 AND tenant_id = $2"
    var user User
    err := s.db.QueryRowContext(ctx, query, userID, tenantID).Scan(&user)
    
    return &user, err
}
```

## API Reference

### Functions

#### `WithTenant(ctx context.Context, tenantID string) context.Context`
Stores a tenant ID in the context.

#### `TenantID(ctx context.Context) (string, bool)`
Extracts tenant ID from context. Returns empty string and false if not present.

#### `WithApplication(ctx context.Context, appID string) context.Context`
Stores an application ID in the context. Empty strings are ignored.

#### `AppID(ctx context.Context) (string, bool)`
Extracts application ID from context.

#### `WithAPIKeyPrefix(ctx context.Context, prefix string) context.Context`
Stores an API key prefix for audit tracking.

#### `APIKeyActor(ctx context.Context) (string, bool)`
Extracts API key prefix (actor) from context.

#### `WithTenantAuthValues(ctx context.Context, values TenantAuthValues) context.Context`
Stores combined authentication values in context.

#### `TenantAuth(ctx context.Context) (TenantAuthValues, bool)`
Extracts combined auth values. Falls back to individual extraction if combined values not set.

### Types

#### `TenantAuthValues`

```go
type TenantAuthValues struct {
    TenantID  string     // Tenant identifier
    AppID     string     // Application identifier (optional)
    Prefix    string     // API key prefix for audit trails
    LastUsed  *time.Time // Last time the API key was used
    CreatedAt *time.Time // When the API key was created
}
```

## Best Practices

1. **Always check the boolean return value** - Context values might not be present
2. **Set tenant early in request lifecycle** - Typically in authentication middleware
3. **Use combined values for efficiency** - `WithTenantAuthValues` is more efficient than multiple calls
4. **Don't use for request-scoped data** - Only for cross-cutting concerns like tenant/app identity
5. **Document context requirements** - Make it clear when functions expect tenant context

## Testing

```bash
cd contextx
go test -v
```

## License

MIT License
