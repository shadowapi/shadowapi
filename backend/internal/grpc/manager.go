package grpc

import (
	"log/slog"
	"sync"

	workerv1 "github.com/shadowapi/shadowapi/backend/pkg/proto/worker/v1"
)

// WorkerConnection represents a connected worker
type WorkerConnection struct {
	ID         string
	Name       string
	Stream     workerv1.WorkerService_ConnectServer
	Status     workerv1.WorkerStatus
	ActiveJobs int32
	Capacity   int32
	IsGlobal   bool
	Workspaces []string
	mu         sync.RWMutex
}

// UpdateStatus updates the worker's status from a heartbeat
func (c *WorkerConnection) UpdateStatus(hb *workerv1.Heartbeat) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.Status = hb.Status
	c.ActiveJobs = hb.ActiveJobs
	c.Capacity = hb.Capacity
}

// WorkerManager tracks connected workers
type WorkerManager struct {
	log     *slog.Logger
	workers map[string]*WorkerConnection
	mu      sync.RWMutex
}

// NewWorkerManager creates a new worker manager
func NewWorkerManager(log *slog.Logger) *WorkerManager {
	return &WorkerManager{
		log:     log,
		workers: make(map[string]*WorkerConnection),
	}
}

// Register adds a worker connection to the manager
func (m *WorkerManager) Register(id, name string, stream workerv1.WorkerService_ConnectServer, isGlobal bool, workspaces []string) *WorkerConnection {
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
	m.log.Info("worker registered", "id", id, "name", name, "is_global", isGlobal)
	return conn
}

// Unregister removes a worker connection from the manager
func (m *WorkerManager) Unregister(id string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if w, ok := m.workers[id]; ok {
		m.log.Info("worker unregistered", "id", id, "name", w.Name)
		delete(m.workers, id)
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

// Count returns the number of connected workers
func (m *WorkerManager) Count() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.workers)
}
