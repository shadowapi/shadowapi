package filters

import (
	"strings"

	"github.com/shadowapi/shadowapi/backend/pkg/api"
)

type SyncPolicyFilter struct {
	AllowedDomains []string
}

func NewSyncPolicyFilter(allowed []string) *SyncPolicyFilter {
	return &SyncPolicyFilter{AllowedDomains: allowed}
}

func (spf *SyncPolicyFilter) Apply(message *api.Message) bool {
	// For demonstration, check if sender’s email ends with one of the allowed domains.
	sender := message.Sender
	for _, d := range spf.AllowedDomains {
		if strings.HasSuffix(sender, d) {
			return true
		}
	}
	return false
}
