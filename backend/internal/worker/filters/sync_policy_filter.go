package filters

import (
	"context"
	"strings"

	"github.com/shadowapi/shadowapi/backend/pkg/api"
	"log/slog"
)

// SyncPolicyFilter filters messages using an API sync policy.
type SyncPolicyFilter struct {
	policy api.SyncPolicy
	log    *slog.Logger
}

// NewSyncPolicyFilter creates a new SyncPolicyFilter using the provided API sync policy.
func NewSyncPolicyFilter(policy api.SyncPolicy, log *slog.Logger) *SyncPolicyFilter {
	return &SyncPolicyFilter{
		policy: policy,
		log:    log,
	}
}

// Apply returns true if the message is allowed under the sync policy.
// If SyncAll is true, the filter allows all messages.
// Otherwise, the sender is checked against the blocklist and exclude list.
func (spf *SyncPolicyFilter) Apply(ctx context.Context, message *api.Message) bool {
	// If SyncAll is true, allow all messages.
	if spf.policy.GetSyncAll().Or(true) {
		return true
	}
	sender := message.Sender
	// Block message if sender matches any blocked pattern.
	for _, blocked := range spf.policy.GetBlocklist() {
		if strings.Contains(sender, blocked) {
			spf.log.Info("Message blocked by sync policy", "sender", sender, "blocked", blocked)
			return false
		}
	}
	// Exclude message if sender matches any excluded pattern.
	for _, excluded := range spf.policy.GetExcludeList() {
		if strings.Contains(sender, excluded) {
			spf.log.Info("Message excluded by sync policy", "sender", sender, "excluded", excluded)
			return false
		}
	}
	return true
}
