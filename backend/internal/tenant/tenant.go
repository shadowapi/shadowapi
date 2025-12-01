package tenant

import (
	"context"
	"strings"
)

type contextKey string

const TenantContextKey contextKey = "tenant"

// Tenant represents the current tenant context
type Tenant struct {
	UUID        string
	Name        string
	DisplayName string
	IsEnabled   bool
}

// FromContext extracts tenant from context
func FromContext(ctx context.Context) (*Tenant, bool) {
	t, ok := ctx.Value(TenantContextKey).(*Tenant)
	return t, ok
}

// MustFromContext extracts tenant from context or panics
func MustFromContext(ctx context.Context) *Tenant {
	t, ok := FromContext(ctx)
	if !ok {
		panic("tenant not found in context")
	}
	return t
}

// WithTenant returns a new context with the tenant set
func WithTenant(ctx context.Context, t *Tenant) context.Context {
	return context.WithValue(ctx, TenantContextKey, t)
}

// ExtractSubdomain extracts tenant subdomain from host header.
// For example, "acme.localtest.me" with baseDomain "localtest.me" returns "acme".
// Returns empty string if host is the root domain or doesn't match the base domain.
func ExtractSubdomain(host, baseDomain string) string {
	// Remove port if present
	if idx := strings.Index(host, ":"); idx != -1 {
		host = host[:idx]
	}

	// Check if it's the root domain
	if host == baseDomain {
		return ""
	}

	// Check if host ends with .baseDomain
	suffix := "." + baseDomain
	if !strings.HasSuffix(host, suffix) {
		return ""
	}

	// Extract subdomain
	subdomain := strings.TrimSuffix(host, suffix)

	// Don't allow nested subdomains (e.g., "foo.bar.localtest.me")
	if strings.Contains(subdomain, ".") {
		return ""
	}

	return subdomain
}
