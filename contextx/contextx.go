package contextx

import (
	"context"
	"time"
)

type tenantKey struct{}
type applicationKey struct{}
type apiKeyPrefixKey struct{}
type tenantAppValuesKey struct{}

// TenantAuthValues holds authentication context values for multi-tenant applications.
type TenantAuthValues struct {
	TenantID  string     // Tenant identifier
	AppID     string     // Application identifier (optional)
	Prefix    string     // API key prefix for audit trails
	LastUsed  *time.Time // Last time the API key was used
	CreatedAt *time.Time // When the API key was created
}

// WithTenant returns a context containing a tenant ID.
func WithTenant(ctx context.Context, tenantID string) context.Context {
	return context.WithValue(ctx, tenantKey{}, tenantID)
}

// TenantID extracts tenant ID from context if present.
func TenantID(ctx context.Context) (string, bool) {
	v := ctx.Value(tenantKey{})
	if v == nil {
		return "", false
	}
	id, ok := v.(string)
	return id, ok
}

// WithApplication stores an application ID in context.
func WithApplication(ctx context.Context, appID string) context.Context {
	if appID == "" {
		return ctx
	}
	return context.WithValue(ctx, applicationKey{}, appID)
}

// AppID extracts application ID from context if present.
func AppID(ctx context.Context) (string, bool) {
	v := ctx.Value(applicationKey{})
	if v == nil {
		return "", false
	}
	id, ok := v.(string)
	return id, ok
}

// WithAPIKeyPrefix stores an API key prefix (for audit attribution).
func WithAPIKeyPrefix(ctx context.Context, prefix string) context.Context {
	return context.WithValue(ctx, apiKeyPrefixKey{}, prefix)
}

// APIKeyActor returns the API key prefix (actor) if present.
func APIKeyActor(ctx context.Context) (string, bool) {
	v := ctx.Value(apiKeyPrefixKey{})
	if v == nil {
		return "", false
	}
	s, ok := v.(string)
	return s, ok
}

// WithTenantAuthValues stores combined tenant and application auth values in context.
func WithTenantAuthValues(ctx context.Context, values TenantAuthValues) context.Context {
	return context.WithValue(ctx, tenantAppValuesKey{}, values)
}

// TenantAuth extracts combined tenant and application auth values from context.
// Falls back to individual extraction if combined values not found.
func TenantAuth(ctx context.Context) (TenantAuthValues, bool) {
	var result TenantAuthValues

	// Check for combined values first
	tenantAppValues, ok := ctx.Value(tenantAppValuesKey{}).(TenantAuthValues)
	if ok {
		return tenantAppValues, true
	}

	// Fallback to individual extraction
	tenantID, ok := TenantID(ctx)
	if !ok {
		return result, false
	}
	result.TenantID = tenantID
	appID, _ := AppID(ctx)
	result.AppID = appID
	
	return result, true
}
