package manager

import (
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

// GetStatus returns the current worker status
func (c *WorkerConnection) GetStatus() workerv1.WorkerStatus {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.Status
}

// GetActiveJobs returns the number of active jobs
func (c *WorkerConnection) GetActiveJobs() int32 {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.ActiveJobs
}

// GetCapacity returns the worker's capacity
func (c *WorkerConnection) GetCapacity() int32 {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.Capacity
}

// CanAcceptJob checks if the worker can accept another job
func (c *WorkerConnection) CanAcceptJob() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.Status == workerv1.WorkerStatus_WORKER_STATUS_IDLE ||
		(c.Status == workerv1.WorkerStatus_WORKER_STATUS_BUSY && c.ActiveJobs < c.Capacity)
}

// SendJob sends a job assignment to the worker
func (c *WorkerConnection) SendJob(job *workerv1.JobAssignment) error {
	return c.Stream.Send(&workerv1.ServerMessage{
		Payload: &workerv1.ServerMessage_JobAssignment{
			JobAssignment: job,
		},
	})
}

// SendDisconnect sends a disconnect request to the worker
func (c *WorkerConnection) SendDisconnect(reason string) error {
	return c.Stream.Send(&workerv1.ServerMessage{
		Payload: &workerv1.ServerMessage_Disconnect{
			Disconnect: &workerv1.Disconnect{
				Reason: reason,
			},
		},
	})
}
