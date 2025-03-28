// Central registration of all workers
package registry

import (
	"fmt"
	"github.com/shadowapi/shadowapi/backend/internal/worker/types"
	"sync"
)

// Constants for our worker stream and job subjects.
const (
	WorkerStream              = "worker"
	WorkerSubject             = "worker.jobs"
	WorkerSubjectTokenRefresh = WorkerSubject + ".scheduleTokenRefresh"
	WorkerSubjectEmailSync    = WorkerSubject + ".sync.email"
	WorkerSubjectEmailFetch   = WorkerSubject + ".sync.emailFetch"
)

var (
	jobRegistry   = make(map[string]types.JobFactory)
	jobRegistryMu sync.RWMutex

	RegistrySubjects = []string{
		WorkerSubjectTokenRefresh,
		WorkerSubjectEmailSync,
		WorkerSubjectEmailFetch,
	}
)

// RegisterJob registers a job factory for a given subject.
func RegisterJob(subject string, factory types.JobFactory) {
	jobRegistryMu.Lock()
	defer jobRegistryMu.Unlock()
	jobRegistry[subject] = factory
}

// CreateJob returns a new job instance for a given subject.
func CreateJob(subject string, data []byte) (types.Job, error) {
	jobRegistryMu.RLock()
	factory, ok := jobRegistry[subject]
	jobRegistryMu.RUnlock()
	if !ok {
		return nil, fmt.Errorf("no job registered for subject %s", subject)
	}
	return factory(data)
}
