package extractors

import (
	"encoding/json"

	"github.com/shadowapi/shadowapi/backend/pkg/api"
)

type ContactExtractor struct{}

func NewContactExtractor() *ContactExtractor {
	return &ContactExtractor{}
}

// ExtractContact parses the message body (assumed JSON) into a Contact.
func (ce *ContactExtractor) ExtractContact(message *api.Message) (*api.Contact, error) {
	var contact api.Contact
	if err := json.Unmarshal([]byte(message.Body), &contact); err != nil {
		return nil, err
	}
	return &contact, nil
}
