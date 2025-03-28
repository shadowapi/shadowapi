// Central registration of all workers
package worker

import (
	"fmt"
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

// JobFactory is a function that builds a Job from a message’s raw data.
type JobFactory func(data []byte) (Job, error)

var (
	jobRegistry   = make(map[string]JobFactory)
	jobRegistryMu sync.RWMutex

	RegistrySubjects = []string{
		WorkerSubjectTokenRefresh,
		WorkerSubjectEmailSync,
		WorkerSubjectEmailFetch,
	}
)

// RegisterJob registers a job factory for a given subject.
func RegisterJob(subject string, factory JobFactory) {
	jobRegistryMu.Lock()
	defer jobRegistryMu.Unlock()
	jobRegistry[subject] = factory
}

// CreateJob returns a new job instance for a given subject.
func CreateJob(subject string, data []byte) (Job, error) {
	jobRegistryMu.RLock()
	factory, ok := jobRegistry[subject]
	jobRegistryMu.RUnlock()
	if !ok {
		return nil, fmt.Errorf("no job registered for subject %s", subject)
	}
	return factory(data)
}
