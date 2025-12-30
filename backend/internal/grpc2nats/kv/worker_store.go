package kv

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	"github.com/nats-io/nats.go/jetstream"
	"github.com/samber/do/v2"

	"github.com/shadowapi/shadowapi/backend/internal/grpc2nats/config"
	"github.com/shadowapi/shadowapi/backend/internal/grpc2nats/natsconn"
)

// WorkerState represents the state of a worker stored in NATS KV
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

// WorkerStore manages worker state in NATS KV store
type WorkerStore struct {
	kv         jetstream.KeyValue
	cfg        *config.Config
	log        *slog.Logger
	instanceID string
}

// Provide creates a WorkerStore for dependency injection
func Provide(i do.Injector) (*WorkerStore, error) {
	cfg := do.MustInvoke[*config.Config](i)
	conn := do.MustInvoke[*natsconn.Connection](i)
	log := do.MustInvoke[*slog.Logger](i).With("component", "kv-store")

	ctx := context.Background()
	bucketName := cfg.NATS.Prefix + "-" + cfg.KV.WorkerBucket

	log.Debug("creating KV bucket", "bucket", bucketName)

	kv, err := conn.JS.CreateOrUpdateKeyValue(ctx, jetstream.KeyValueConfig{
		Bucket:      bucketName,
		Description: "Worker connection state for grpc2nats",
		TTL:         5 * time.Minute, // Workers must refresh within 5 minutes
	})
	if err != nil {
		log.Error("failed to create KV bucket", "error", err)
		return nil, err
	}

	log.Info("KV store initialized", "bucket", bucketName)

	return &WorkerStore{
		kv:         kv,
		cfg:        cfg,
		log:        log,
		instanceID: cfg.InstanceID,
	}, nil
}

// Put stores a worker's state
func (s *WorkerStore) Put(ctx context.Context, state *WorkerState) error {
	state.InstanceID = s.instanceID
	state.LastSeen = time.Now().UTC()

	data, err := json.Marshal(state)
	if err != nil {
		return err
	}

	_, err = s.kv.Put(ctx, state.WorkerID, data)
	if err != nil {
		s.log.Error("failed to put worker state", "worker_id", state.WorkerID, "error", err)
		return err
	}

	s.log.Debug("worker state updated", "worker_id", state.WorkerID)
	return nil
}

// Get retrieves a worker's state
func (s *WorkerStore) Get(ctx context.Context, workerID string) (*WorkerState, error) {
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

// Delete removes a worker's state
func (s *WorkerStore) Delete(ctx context.Context, workerID string) error {
	err := s.kv.Delete(ctx, workerID)
	if err != nil {
		s.log.Error("failed to delete worker state", "worker_id", workerID, "error", err)
		return err
	}

	s.log.Debug("worker state deleted", "worker_id", workerID)
	return nil
}

// ListAll returns all worker states
func (s *WorkerStore) ListAll(ctx context.Context) ([]*WorkerState, error) {
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

// ListByWorkspace returns workers that can process jobs for a workspace
func (s *WorkerStore) ListByWorkspace(ctx context.Context, workspaceSlug string) ([]*WorkerState, error) {
	all, err := s.ListAll(ctx)
	if err != nil {
		return nil, err
	}

	var result []*WorkerState
	for _, state := range all {
		if state.IsGlobal {
			result = append(result, state)
			continue
		}
		for _, ws := range state.Workspaces {
			if ws == workspaceSlug {
				result = append(result, state)
				break
			}
		}
	}

	return result, nil
}

// ListByInstance returns workers connected to a specific grpc2nats instance
func (s *WorkerStore) ListByInstance(ctx context.Context, instanceID string) ([]*WorkerState, error) {
	all, err := s.ListAll(ctx)
	if err != nil {
		return nil, err
	}

	var result []*WorkerState
	for _, state := range all {
		if state.InstanceID == instanceID {
			result = append(result, state)
		}
	}

	return result, nil
}

// CleanupInstance removes all worker states for an instance
func (s *WorkerStore) CleanupInstance(ctx context.Context, instanceID string) error {
	workers, err := s.ListByInstance(ctx, instanceID)
	if err != nil {
		return err
	}

	for _, w := range workers {
		if err := s.Delete(ctx, w.WorkerID); err != nil {
			s.log.Warn("failed to delete worker during cleanup", "worker_id", w.WorkerID, "error", err)
		}
	}

	s.log.Info("cleaned up workers for instance", "instance_id", instanceID, "count", len(workers))
	return nil
}
