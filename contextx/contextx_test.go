package contextx

import (
	"context"
	"testing"
	"time"
)

func TestWithTenantAndTenantID(t *testing.T) {
	ctx := context.Background()
	tenantID := "tenant-123"

	ctx = WithTenant(ctx, tenantID)
	extracted, ok := TenantID(ctx)

	if !ok {
		t.Fatal("expected tenant ID to be present")
	}
	if extracted != tenantID {
		t.Fatalf("expected %s, got %s", tenantID, extracted)
	}
}

func TestTenantIDNotPresent(t *testing.T) {
	ctx := context.Background()
	_, ok := TenantID(ctx)

	if ok {
		t.Fatal("expected tenant ID to not be present")
	}
}

func TestWithApplicationAndAppID(t *testing.T) {
	ctx := context.Background()
	appID := "app-456"

	ctx = WithApplication(ctx, appID)
	extracted, ok := AppID(ctx)

	if !ok {
		t.Fatal("expected app ID to be present")
	}
	if extracted != appID {
		t.Fatalf("expected %s, got %s", appID, extracted)
	}
}

func TestWithApplicationEmptyString(t *testing.T) {
	ctx := context.Background()
	ctx = WithApplication(ctx, "")

	_, ok := AppID(ctx)
	if ok {
		t.Fatal("expected empty app ID to not be stored")
	}
}

func TestWithAPIKeyPrefix(t *testing.T) {
	ctx := context.Background()
	prefix := "sk_live_"

	ctx = WithAPIKeyPrefix(ctx, prefix)
	extracted, ok := APIKeyActor(ctx)

	if !ok {
		t.Fatal("expected API key prefix to be present")
	}
	if extracted != prefix {
		t.Fatalf("expected %s, got %s", prefix, extracted)
	}
}

func TestWithTenantAuthValues(t *testing.T) {
	ctx := context.Background()
	now := time.Now()
	
	values := TenantAuthValues{
		TenantID:  "tenant-123",
		AppID:     "app-456",
		Prefix:    "sk_test_",
		LastUsed:  &now,
		CreatedAt: &now,
	}

	ctx = WithTenantAuthValues(ctx, values)
	extracted, ok := TenantAuth(ctx)

	if !ok {
		t.Fatal("expected tenant auth values to be present")
	}
	if extracted.TenantID != values.TenantID {
		t.Fatalf("expected tenant ID %s, got %s", values.TenantID, extracted.TenantID)
	}
	if extracted.AppID != values.AppID {
		t.Fatalf("expected app ID %s, got %s", values.AppID, extracted.AppID)
	}
	if extracted.Prefix != values.Prefix {
		t.Fatalf("expected prefix %s, got %s", values.Prefix, extracted.Prefix)
	}
}

func TestTenantAuthFallback(t *testing.T) {
	ctx := context.Background()
	
	// Set individual values without using WithTenantAuthValues
	ctx = WithTenant(ctx, "tenant-789")
	ctx = WithApplication(ctx, "app-101")

	extracted, ok := TenantAuth(ctx)

	if !ok {
		t.Fatal("expected tenant auth to work with fallback")
	}
	if extracted.TenantID != "tenant-789" {
		t.Fatalf("expected tenant ID tenant-789, got %s", extracted.TenantID)
	}
	if extracted.AppID != "app-101" {
		t.Fatalf("expected app ID app-101, got %s", extracted.AppID)
	}
}

func TestTenantAuthNoTenant(t *testing.T) {
	ctx := context.Background()
	
	_, ok := TenantAuth(ctx)
	if ok {
		t.Fatal("expected tenant auth to fail when no tenant ID")
	}
}
