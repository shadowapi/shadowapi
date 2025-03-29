// Shared types and interfaces
// internal/worker/types.go
package types

import (
	"context"
	"fmt"
	"github.com/shadowapi/shadowapi/backend/pkg/api"
	"time"
)

// JobNotReadyError signals that a job is not ready to be processed.
// It includes a Delay field that indicates how long to wait before retrying.
type JobNotReadyError struct {
	Delay time.Duration
}

func (e JobNotReadyError) Error() string {
	return fmt.Sprintf("job not ready, try again after %s", e.Delay)
}

// Job represents a unit of work to be executed.
type Job interface {
	// Execute runs the job logic.
	Execute(ctx context.Context) error
}

// Worker represents any entity capable of processing a job.
type Worker interface {
	// Work processes a given job.
	Work(ctx context.Context, job Job) error
}

// Extractor extracts information from a message.
type Extractor interface {
	// ExtractContact extracts a Contact from the given message.
	ExtractContact(message *api.Message) (*api.Contact, error)
}

// Filter determines whether a message meets some criteria.
type Filter interface {
	Apply(ctx context.Context, message *api.Message) bool
}

// Storage saves message and attachment data.
type Storage interface {
	// SaveMessage persists a message.
	SaveMessage(ctx context.Context, message *api.Message) error
	// SaveAttachment persists a file attachment.
	SaveAttachment(ctx context.Context, file *api.FileObject) error
}

// Pipeline represents a chain of processing steps to run on a message.
type Pipeline interface {
	// Run executes the pipeline steps.
	Run(ctx context.Context, message *api.Message) error
}

// JobFactory is a function that builds a Job from a message’s raw data.
type JobFactory func(data []byte) (Job, error)
