package policies

import (
	"context"
	"github.com/ory/ladon"
	"github.com/ory/ladon/manager/memory"
)

// NewPolicyManager returns an in–memory ladon manager with some default policies.
func NewPolicyManager() *MemoryPolicies {
	// The ory/ladon memory manager is in github.com/ory/ladon/manager/memory
	manager := memory.NewMemoryManager()

	// Optionally, create or add any default policies:
	policies := []ladon.Policy{
		&ladon.DefaultPolicy{
			ID:          "allow-sample",
			Description: "Example policy that allows 'sync' action",
			Subjects:    []string{"<.*>"},
			Resources:   []string{"email:<[0-9a-zA-Z\\-]+>"},
			Actions:     []string{"sync"},
			Effect:      ladon.AllowAccess,
		},
	}
	ctx := context.Background()
	for _, pol := range policies {
		_ = manager.Create(ctx, pol)
	}

	// Construct the Ladon instance, with the memory manager
	ladonEnforcer := &ladon.Ladon{
		Manager: manager,
		// optionally set a custom matcher or audit logger
	}

	return &MemoryPolicies{
		Ladon: ladonEnforcer,
	}
}
