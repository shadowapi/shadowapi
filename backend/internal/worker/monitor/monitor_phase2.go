package monitor

import (
	"github.com/prometheus/client_golang/prometheus"
	"log/slog"
)

var (
	phase2JobStartCounter = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "worker_phase2_job_start_total",
		Help: "Total number of phase2 job starts",
	})
	phase2JobEndCounter = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "worker_phase2_job_end_total",
		Help: "Total number of phase2 job ends",
	})
)

func init() {
	prometheus.MustRegister(phase2JobStartCounter, phase2JobEndCounter)
}

// Phase2Monitor shows sample Phase2 monitoring with Prometheus metrics and potential tracing.
type Phase2Monitor struct {
	log *slog.Logger
}

// NewPhase2Monitor creates a new Phase2Monitor.
func NewPhase2Monitor(log *slog.Logger) *Phase2Monitor {
	return &Phase2Monitor{
		log: log,
	}
}

// RecordJobStart records a job start event with extended metrics.
func (p *Phase2Monitor) RecordJobStart(jobID, subject string) {
	p.log.Debug("Phase2Monitor: job start", "jobID", jobID, "subject", subject)
	phase2JobStartCounter.Inc()
	// Here you could also start an OpenTelemetry trace span.
}

// RecordJobEnd records a job end event and updates metrics.
func (p *Phase2Monitor) RecordJobEnd(jobID, subject, finalStatus string) {
	p.log.Debug("Phase2Monitor: job end", "jobID", jobID, "subject", subject, "status", finalStatus)
	phase2JobEndCounter.Inc()
	// Here you could end the trace span or add additional tracing information.
}
