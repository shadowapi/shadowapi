package session

import (
	"context"
)

type identityKey string

type Identity struct {
	ID string `json:"id"`
}

// WithIdentity stores the identity in the context
func WithIdentity(ctx context.Context, id Identity) context.Context {
	return context.WithValue(ctx, identityKey("identityKey"), id)
}

// GetIdentity retrieves the identity from the context
func GetIdentity(ctx context.Context) (Identity, bool) {
	id, ok := ctx.Value(identityKey("identityKey")).(Identity)
	return id, ok
}
