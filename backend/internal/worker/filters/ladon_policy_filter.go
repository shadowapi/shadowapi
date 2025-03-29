package filters

import (
	"context"
	"github.com/shadowapi/shadowapi/backend/internal/worker/policies"
	"github.com/shadowapi/shadowapi/backend/pkg/api"
	"log/slog"
)

// LadonPolicyFilter checks if user has permission to sync the message's resource
type LadonPolicyFilter struct {
	log       *slog.Logger
	polMgr    *policies.MemoryPolicies
	userID    string
	accountID string
}

func NewLadonPolicyFilter(
	log *slog.Logger,
	polMgr *policies.MemoryPolicies,
	userID string,
	accountID string,
) *LadonPolicyFilter {
	return &LadonPolicyFilter{log: log, polMgr: polMgr, userID: userID, accountID: accountID}
}

func (f *LadonPolicyFilter) Apply(ctx context.Context, message *api.Message) bool {
	// For example, resource = "email:<accountID>"
	resource := "email:" + f.accountID
	// Action = "sync"
	if err := f.polMgr.Check(ctx, f.userID, resource, "sync"); err != nil {
		f.log.Warn("Policy denied sync", "user", f.userID, "account", f.accountID, "err", err)
		return false
	}
	return true
}
