// Package jobstore provides NATS JetStream KV storage for temporary job results
package jobstore

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	"github.com/nats-io/nats.go/jetstream"
	"github.com/samber/do/v2"

	"github.com/shadowapi/shadowapi/backend/internal/queue"
)

const (
	// DefaultTTL is the time-to-live for job results (5 minutes)
	DefaultTTL = 5 * time.Minute

	// BucketSuffix is appended to the queue prefix for the bucket name
	BucketSuffix = "job-results"
)

// JobResult represents a stored job result in the KV store
type JobResult struct {
	UUID         string    `json:"uuid"`
	ResourceType string    `json:"resource_type"` // "email_oauth", "postgres"
	ResourceUUID string    `json:"resource_uuid"`
	Status       string    `json:"status"` // pending, running, completed, failed
	Result       []byte    `json:"result,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
	CompletedAt  time.Time `json:"completed_at,omitempty"`
}

// Store manages job results in NATS KV
type Store struct {
	kv  jetstream.KeyValue
	log *slog.Logger
}

// Provide creates a Store for dependency injection
func Provide(i do.Injector) (*Store, error) {
	q := do.MustInvoke[*queue.Queue](i)
	log := do.MustInvoke[*slog.Logger](i).With("component", "jobstore")

	ctx := context.Background()
	bucketName := q.Prefix() + "-" + BucketSuffix

	log.Debug("creating job results KV bucket", "bucket", bucketName)

	kv, err := q.CreateOrUpdateKeyValue(ctx, jetstream.KeyValueConfig{
		Bucket:      bucketName,
		Description: "Temporary job results storage",
		TTL:         DefaultTTL,
	})
	if err != nil {
		log.Error("failed to create KV bucket", "error", err)
		return nil, err
	}

	log.Info("job results KV store initialized", "bucket", bucketName, "ttl", DefaultTTL)

	return &Store{
		kv:  kv,
		log: log,
	}, nil
}

// Put stores a job result
func (s *Store) Put(ctx context.Context, job *JobResult) error {
	data, err := json.Marshal(job)
	if err != nil {
		return err
	}

	_, err = s.kv.Put(ctx, job.UUID, data)
	if err != nil {
		s.log.Error("failed to put job result", "job_uuid", job.UUID, "error", err)
		return err
	}

	s.log.Debug("job result stored", "job_uuid", job.UUID, "status", job.Status)
	return nil
}

// Get retrieves a job result by UUID
func (s *Store) Get(ctx context.Context, uuid string) (*JobResult, error) {
	entry, err := s.kv.Get(ctx, uuid)
	if err != nil {
		return nil, err
	}

	var job JobResult
	if err := json.Unmarshal(entry.Value(), &job); err != nil {
		return nil, err
	}

	return &job, nil
}

// Delete removes a job result
func (s *Store) Delete(ctx context.Context, uuid string) error {
	err := s.kv.Delete(ctx, uuid)
	if err != nil {
		s.log.Error("failed to delete job result", "job_uuid", uuid, "error", err)
		return err
	}

	s.log.Debug("job result deleted", "job_uuid", uuid)
	return nil
}
