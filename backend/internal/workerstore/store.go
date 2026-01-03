// Package workerstore provides read access to worker state from NATS KV
package workerstore

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
	// BucketSuffix is appended to the queue prefix for the bucket name
	// Must match grpc2nats KV bucket naming: {prefix}-workers
	BucketSuffix = "workers"
)

// WorkerState represents the state of a worker stored in NATS KV
// This struct matches the one defined in grpc2nats/kv/worker_store.go
type WorkerState struct {
	WorkerID    string    `json:"worker_id"`
	WorkerName  string    `json:"worker_name"`
	InstanceID  string    `json:"instance_id"`
	Status      string    `json:"status"`
	ActiveJobs  int32     `json:"active_jobs"`
	Capacity    int32     `json:"capacity"`
	IsGlobal    bool      `json:"is_global"`
	Workspaces  []string  `json:"workspaces"`
	ConnectedAt time.Time `json:"connected_at"`
	LastSeen    time.Time `json:"last_seen"`
}

// Store provides read access to worker state from NATS KV
type Store struct {
	kv  jetstream.KeyValue
	log *slog.Logger
}

// Provide creates a Store for dependency injection
func Provide(i do.Injector) (*Store, error) {
	q := do.MustInvoke[*queue.Queue](i)
	log := do.MustInvoke[*slog.Logger](i).With("component", "workerstore")

	ctx := context.Background()
	bucketName := q.Prefix() + "-" + BucketSuffix

	log.Debug("accessing worker state KV bucket", "bucket", bucketName)

	// Use CreateOrUpdateKeyValue to ensure bucket exists
	// This is safe even if grpc2nats already created it
	kv, err := q.CreateOrUpdateKeyValue(ctx, jetstream.KeyValueConfig{
		Bucket:      bucketName,
		Description: "Worker connection state",
		TTL:         5 * time.Minute, // Same as grpc2nats
	})
	if err != nil {
		log.Error("failed to access KV bucket", "error", err)
		return nil, err
	}

	log.Info("worker state KV store initialized", "bucket", bucketName)

	return &Store{
		kv:  kv,
		log: log,
	}, nil
}

// Get retrieves a worker's state by worker UUID
func (s *Store) Get(ctx context.Context, workerID string) (*WorkerState, error) {
	entry, err := s.kv.Get(ctx, workerID)
	if err != nil {
		return nil, err
	}

	var state WorkerState
	if err := json.Unmarshal(entry.Value(), &state); err != nil {
		return nil, err
	}

	return &state, nil
}

// ListAll returns all worker states
func (s *Store) ListAll(ctx context.Context) ([]*WorkerState, error) {
	keys, err := s.kv.Keys(ctx)
	if err != nil {
		// If no keys exist, return empty slice
		if err == jetstream.ErrNoKeysFound {
			return []*WorkerState{}, nil
		}
		return nil, err
	}

	var states []*WorkerState
	for _, key := range keys {
		state, err := s.Get(ctx, key)
		if err != nil {
			s.log.Warn("failed to get worker state", "key", key, "error", err)
			continue
		}
		states = append(states, state)
	}

	return states, nil
}
