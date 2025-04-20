// Central registration of all workers
package registry

import (
	"fmt"
	"github.com/shadowapi/shadowapi/backend/internal/worker/types"
	"sync"
)

// Constants for our worker stream and job subjects.
const (
	WorkerStream                    = "worker"
	WorkerSubject                   = "worker.jobs"
	WorkerSubjectTokenRefresh       = WorkerSubject + ".scheduleTokenRefresh"
	WorkerSubjectEmailOAuthFetch    = WorkerSubject + ".emailOAuthFetch"
	WorkerSubjectEmailApplyPipeline = WorkerSubject + ".emailApplyPipeline"
)

var (
	jobRegistry   = make(map[string]types.JobFactory)
	jobRegistryMu sync.RWMutex

	RegistrySubjects = []string{
		WorkerSubjectTokenRefresh,
		WorkerSubjectEmailOAuthFetch, // enable scheduled Gmail OAuth2 fetch jobs
		WorkerSubjectEmailApplyPipeline,
	}
)

// RegisterJob registers a job factory for a given subject.
func RegisterJob(subject string, factory types.JobFactory) {
	jobRegistryMu.Lock()
	defer jobRegistryMu.Unlock()
	jobRegistry[subject] = factory
}

// CreateJob returns a new job instance for a given subject.
func CreateJob(subject string, natsJobID string, data []byte) (types.Job, error) {
	// TODO @reactima decide if it's worth to handle natsJobID
	jobRegistryMu.RLock()
	factory, ok := jobRegistry[subject]
	jobRegistryMu.RUnlock()
	if !ok {
		return nil, fmt.Errorf("no job registered for subject %s", subject)
	}
	return factory(data)
}
