package handler

import (
	"context"
	"encoding/json"

	"github.com/go-faster/jx"

	"github.com/shadowapi/shadowapi/backend/internal/worker/subjects"
	"github.com/shadowapi/shadowapi/backend/internal/workspace"
	"github.com/shadowapi/shadowapi/backend/pkg/api"
)

// NatsMessagesList returns the last messages from the NATS data stream
// GET /nats/messages
func (h *Handler) NatsMessagesList(ctx context.Context, params api.NatsMessagesListParams) (api.NatsMessagesListRes, error) {
	log := h.log.With("handler", "NatsMessagesList")

	// Get workspace from context
	workspaceSlug := workspace.GetWorkspaceSlug(ctx)
	if workspaceSlug == "" {
		workspaceSlug = "internal"
	}

	// Determine limit
	limit := 50
	if params.Limit.IsSet() && params.Limit.Value > 0 && params.Limit.Value <= 500 {
		limit = params.Limit.Value
	}

	// Build subject filter for this workspace's messages
	subjectFilter := subjects.DataSubject(workspaceSlug, "messages")

	log.Debug("fetching NATS messages",
		"workspace", workspaceSlug,
		"subject", subjectFilter,
		"limit", limit,
	)

	// Fetch messages from NATS stream
	streamMessages, err := h.queue.GetLastMessages(ctx, subjects.DataStreamName, subjectFilter, limit)
	if err != nil {
		log.Warn("failed to get messages from NATS", "error", err)
		// Return empty list rather than error if stream doesn't exist
		return &api.NatsMessagesList{
			Messages: []api.NatsMessage{},
			Total:    0,
		}, nil
	}

	// Convert to API response
	messages := make([]api.NatsMessage, 0, len(streamMessages))
	for _, msg := range streamMessages {
		natsMsg := api.NatsMessage{
			Sequence:  int(msg.Sequence),
			Subject:   msg.Subject,
			Timestamp: int(msg.Timestamp),
		}

		// Extract job_id from headers if present
		if jobID, ok := msg.Headers["X-Job-ID"]; ok {
			natsMsg.JobID = api.NewOptString(jobID)
		}

		// Parse data as JSON and convert to jx.Raw values
		if len(msg.Data) > 0 {
			var data map[string]json.RawMessage
			if err := json.Unmarshal(msg.Data, &data); err == nil {
				// Convert to NatsMessageData (map[string]jx.Raw)
				natsData := make(api.NatsMessageData, len(data))
				for k, v := range data {
					natsData[k] = jx.Raw(v)
				}
				natsMsg.Data = api.NewOptNatsMessageData(natsData)
			}
		}

		messages = append(messages, natsMsg)
	}

	log.Info("returned NATS messages", "count", len(messages))

	return &api.NatsMessagesList{
		Messages: messages,
		Total:    len(messages),
	}, nil
}
