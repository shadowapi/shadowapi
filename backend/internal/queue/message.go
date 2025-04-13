package queue

import (
	"time"
)

type Msg interface {
	// Subject returns a subject on which a message was published/received.
	Subject() string

	// Data returns the message body.
	Data() []byte

	// Ack acknowledges a message. This tells the server that the message was
	// successfully processed and it can move on to the next message.
	Ack() error

	// Nak negatively acknowledges a message. This tells the server to
	// redeliver the message.
	//
	// Nak does not adhere to AckWait or Backoff configured on the consumer
	// and triggers instant redelivery. For a delayed redelivery, use
	// NakWithDelay.
	Nak() error

	// NakWithDelay negatively acknowledges a message. This tells the server
	// to redeliver the message after the given delay.
	NakWithDelay(delay time.Duration) error

	// InProgress tells the server that this message is being worked on. It
	// resets the redelivery timer on the server.
	InProgress() error

	// Term tells the server to not redeliver this message, regardless of
	// the value of MaxDeliver.
	Term() error
}
