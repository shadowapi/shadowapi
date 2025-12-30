package manager

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/samber/do/v2"

	"github.com/shadowapi/shadowapi/backend/internal/grpc2nats/config"
	"github.com/shadowapi/shadowapi/backend/internal/grpc2nats/kv"
	workerv1 "github.com/shadowapi/shadowapi/backend/pkg/proto/worker/v1"
)

// WorkerManager tracks connected workers and syncs with NATS KV
type WorkerManager struct {
	log        *slog.Logger
	cfg        *config.Config
	kvStore    *kv.WorkerStore
	workers    map[string]*WorkerConnection
	mu         sync.RWMutex
	instanceID string
}

// Provide creates a WorkerManager for dependency injection
func Provide(i do.Injector) (*WorkerManager, error) {
	log := do.MustInvoke[*slog.Logger](i).With("component", "worker-manager")
	cfg := do.MustInvoke[*config.Config](i)
	kvStore := do.MustInvoke[*kv.WorkerStore](i)

	m := &WorkerManager{
		log:        log,
		cfg:        cfg,
		kvStore:    kvStore,
		workers:    make(map[string]*WorkerConnection),
		instanceID: cfg.InstanceID,
	}

	// Clean up any stale entries for this instance on startup
	ctx := context.Background()
	if err := kvStore.CleanupInstance(ctx, cfg.InstanceID); err != nil {
		log.Warn("failed to cleanup stale workers", "error", err)
	}

	return m, nil
}

// Register adds a worker connection to the manager and syncs to KV
func (m *WorkerManager) Register(
	id, name string,
	stream workerv1.WorkerService_ConnectServer,
	isGlobal bool,
	workspaces []string,
) *WorkerConnection {
	m.mu.Lock()
	defer m.mu.Unlock()

	conn := &WorkerConnection{
		ID:         id,
		Name:       name,
		Stream:     stream,
		Status:     workerv1.WorkerStatus_WORKER_STATUS_IDLE,
		IsGlobal:   isGlobal,
		Workspaces: workspaces,
	}
	m.workers[id] = conn

	// Sync to KV store
	ctx := context.Background()
	state := &kv.WorkerState{
		WorkerID:    id,
		WorkerName:  name,
		InstanceID:  m.instanceID,
		Status:      "online",
		IsGlobal:    isGlobal,
		Workspaces:  workspaces,
		ConnectedAt: time.Now().UTC(),
	}
	if err := m.kvStore.Put(ctx, state); err != nil {
		m.log.Warn("failed to sync worker to KV", "worker_id", id, "error", err)
	}

	m.log.Info("worker registered", "id", id, "name", name, "is_global", isGlobal)
	return conn
}

// Unregister removes a worker connection from the manager and KV
func (m *WorkerManager) Unregister(id string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if w, ok := m.workers[id]; ok {
		m.log.Info("worker unregistered", "id", id, "name", w.Name)
		delete(m.workers, id)

		// Remove from KV store
		ctx := context.Background()
		if err := m.kvStore.Delete(ctx, id); err != nil {
			m.log.Warn("failed to delete worker from KV", "worker_id", id, "error", err)
		}
	}
}

// Get returns a worker connection by ID
func (m *WorkerManager) Get(id string) *WorkerConnection {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.workers[id]
}

// GetConnected returns all connected workers
func (m *WorkerManager) GetConnected() []*WorkerConnection {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make([]*WorkerConnection, 0, len(m.workers))
	for _, c := range m.workers {
		result = append(result, c)
	}
	return result
}

// GetByWorkspace returns all workers that can process jobs for a workspace
func (m *WorkerManager) GetByWorkspace(workspaceSlug string) []*WorkerConnection {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var result []*WorkerConnection
	for _, c := range m.workers {
		if c.IsGlobal {
			result = append(result, c)
			continue
		}
		for _, ws := range c.Workspaces {
			if ws == workspaceSlug {
				result = append(result, c)
				break
			}
		}
	}
	return result
}

// GetAvailable returns workers that can accept new jobs for a workspace
func (m *WorkerManager) GetAvailable(workspaceSlug string) []*WorkerConnection {
	workers := m.GetByWorkspace(workspaceSlug)
	var available []*WorkerConnection
	for _, w := range workers {
		if w.CanAcceptJob() {
			available = append(available, w)
		}
	}
	return available
}

// GetGlobalAvailable returns global workers that can accept new jobs
func (m *WorkerManager) GetGlobalAvailable() []*WorkerConnection {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var result []*WorkerConnection
	for _, c := range m.workers {
		if c.IsGlobal && c.CanAcceptJob() {
			result = append(result, c)
		}
	}
	return result
}

// Count returns the number of connected workers
func (m *WorkerManager) Count() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.workers)
}

// UpdateHeartbeat updates worker state from heartbeat and syncs to KV
func (m *WorkerManager) UpdateHeartbeat(id string, hb *workerv1.Heartbeat) {
	conn := m.Get(id)
	if conn == nil {
		return
	}

	conn.UpdateStatus(hb)

	// Sync to KV store
	ctx := context.Background()
	state := &kv.WorkerState{
		WorkerID:   id,
		WorkerName: conn.Name,
		InstanceID: m.instanceID,
		Status:     statusToString(hb.Status),
		ActiveJobs: hb.ActiveJobs,
		Capacity:   hb.Capacity,
		IsGlobal:   conn.IsGlobal,
		Workspaces: conn.Workspaces,
	}
	if err := m.kvStore.Put(ctx, state); err != nil {
		m.log.Warn("failed to sync heartbeat to KV", "worker_id", id, "error", err)
	}
}

// statusToString converts WorkerStatus enum to string
func statusToString(s workerv1.WorkerStatus) string {
	switch s {
	case workerv1.WorkerStatus_WORKER_STATUS_IDLE:
		return "idle"
	case workerv1.WorkerStatus_WORKER_STATUS_BUSY:
		return "busy"
	case workerv1.WorkerStatus_WORKER_STATUS_DRAINING:
		return "draining"
	default:
		return "unknown"
	}
}
