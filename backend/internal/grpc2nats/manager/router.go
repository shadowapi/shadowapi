package manager

import (
	"errors"
	"math/rand"
)

// ErrNoWorkersAvailable is returned when no workers are available for a job
var ErrNoWorkersAvailable = errors.New("no workers available")

// Router handles routing jobs to appropriate workers
type Router struct {
	manager *WorkerManager
}

// NewRouter creates a new Router
func NewRouter(manager *WorkerManager) *Router {
	return &Router{manager: manager}
}

// RouteJob selects a worker to handle a job based on workspace and availability
func (r *Router) RouteJob(workspaceSlug string, isGlobal bool) (*WorkerConnection, error) {
	var candidates []*WorkerConnection

	if isGlobal {
		// For global jobs, prefer global workers but allow any worker
		candidates = r.manager.GetGlobalAvailable()
		if len(candidates) == 0 {
			// Fall back to all available workers if no global ones
			candidates = r.manager.GetAvailable(workspaceSlug)
		}
	} else {
		// For workspace-scoped jobs, get workers for that workspace
		candidates = r.manager.GetAvailable(workspaceSlug)
	}

	if len(candidates) == 0 {
		return nil, ErrNoWorkersAvailable
	}

	// Select worker with least load
	return r.selectLeastLoaded(candidates), nil
}

// selectLeastLoaded picks the worker with the lowest active jobs ratio
func (r *Router) selectLeastLoaded(workers []*WorkerConnection) *WorkerConnection {
	if len(workers) == 0 {
		return nil
	}

	if len(workers) == 1 {
		return workers[0]
	}

	var best *WorkerConnection
	var bestRatio float64 = 2.0 // Start with impossible ratio

	for _, w := range workers {
		activeJobs := w.GetActiveJobs()
		capacity := w.GetCapacity()

		// Avoid division by zero
		if capacity == 0 {
			capacity = 1
		}

		ratio := float64(activeJobs) / float64(capacity)

		// Prefer lower ratio (less loaded)
		if ratio < bestRatio {
			bestRatio = ratio
			best = w
		} else if ratio == bestRatio {
			// If equal ratio, randomly pick one to distribute load
			if rand.Intn(2) == 0 {
				best = w
			}
		}
	}

	return best
}

// RouteToWorker routes a job to a specific worker by ID
func (r *Router) RouteToWorker(workerID string) (*WorkerConnection, error) {
	worker := r.manager.Get(workerID)
	if worker == nil {
		return nil, ErrNoWorkersAvailable
	}

	if !worker.CanAcceptJob() {
		return nil, ErrNoWorkersAvailable
	}

	return worker, nil
}

// GetAllAvailable returns all workers that can accept jobs
func (r *Router) GetAllAvailable() []*WorkerConnection {
	connected := r.manager.GetConnected()
	var available []*WorkerConnection
	for _, w := range connected {
		if w.CanAcceptJob() {
			available = append(available, w)
		}
	}
	return available
}
