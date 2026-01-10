package rbac

import (
	"context"
	"regexp"

	"github.com/ory/ladon"
)

func init() {
	// Register custom conditions with Ladon
	ladon.ConditionFactories[new(WorkspaceCondition).GetName()] = func() ladon.Condition {
		return new(WorkspaceCondition)
	}
}

// WorkspaceCondition checks if request context contains a matching workspace.
// Used for workspace-scoped policies where the policy applies to specific workspaces
// or all workspaces (when Workspaces contains "*").
type WorkspaceCondition struct {
	Workspaces []string `json:"workspaces"`
}

// GetName returns the condition name for Ladon registration.
func (c *WorkspaceCondition) GetName() string {
	return "WorkspaceCondition"
}

// Fulfills checks if the workspace in the request context matches the condition.
func (c *WorkspaceCondition) Fulfills(ctx context.Context, value any, r *ladon.Request) bool {
	workspace, ok := r.Context["workspace"].(string)
	if !ok || workspace == "" {
		return false
	}

	for _, w := range c.Workspaces {
		// Wildcard matches all workspaces
		if w == "*" {
			return true
		}
		// Exact match
		if w == workspace {
			return true
		}
		// Support regex patterns (e.g., "internal|demo")
		if matched, _ := regexp.MatchString("^"+w+"$", workspace); matched {
			return true
		}
	}
	return false
}
