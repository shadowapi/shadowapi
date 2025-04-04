package metrics

import "github.com/prometheus/client_golang/prometheus"

var (
	JobScheduledTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "scheduler_job_scheduled_total",
			Help: "Number of jobs scheduled by the scheduler",
		},
		[]string{"pipeline_uuid", "datasource_uuid"},
	)

	JobExecutedDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "worker_job_executed_duration_seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"pipeline_uuid", "datasource_uuid"},
	)
)

func init() {
	prometheus.MustRegister(JobScheduledTotal, JobExecutedDuration)
}
