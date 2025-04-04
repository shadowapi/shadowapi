package policies

import (
	"context"

	"github.com/ory/ladon"
)

// MemoryPolicies wraps a ladon.Ladon instance and any initialization for your default policies.
type MemoryPolicies struct {
	Ladon *ladon.Ladon
}

// Check checks if the subject can perform the action on resource.
func (mp *MemoryPolicies) Check(ctx context.Context, subject, resource, action string) error {
	return mp.Ladon.IsAllowed(ctx, &ladon.Request{
		Resource: resource,
		Action:   action,
		Subject:  subject,
	})
}
