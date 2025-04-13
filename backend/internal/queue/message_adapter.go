package queue

import (
	"github.com/nats-io/nats.go"
	"time"
)

// HeaderGetter is an optional interface for queue messages to expose header values.
type HeaderGetter interface {
	GetHeader(key string) string
}

// NatsMsgAdapter adapts a *nats.Msg to the queue.Msg interface.
type NatsMsgAdapter struct {
	natsMsg *nats.Msg
}

func (a *NatsMsgAdapter) Subject() string {
	return a.natsMsg.Subject
}

func (a *NatsMsgAdapter) Data() []byte {
	return a.natsMsg.Data
}

func (a *NatsMsgAdapter) Ack() error {
	return a.natsMsg.Ack()
}

func (a *NatsMsgAdapter) NakWithDelay(delay time.Duration) error {
	return a.natsMsg.NakWithDelay(delay)
}

func (a *NatsMsgAdapter) Term() error {
	return a.natsMsg.Term()
}

func (a *NatsMsgAdapter) InProgress() error {
	return a.natsMsg.InProgress()
}

// GetHeader returns the header value for the provided key.
func (a *NatsMsgAdapter) GetHeader(key string) string {
	return a.natsMsg.Header.Get(key)
}
