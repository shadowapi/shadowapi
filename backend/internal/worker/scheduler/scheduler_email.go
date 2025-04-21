package scheduler

import (
	"context"
	"encoding/json"
	"time"

	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/robfig/cron/v3"
	"github.com/shadowapi/shadowapi/backend/internal/converter"
	"github.com/shadowapi/shadowapi/backend/internal/metrics"
	"github.com/shadowapi/shadowapi/backend/internal/queue"
	"github.com/shadowapi/shadowapi/backend/internal/worker/monitor"
	"github.com/shadowapi/shadowapi/backend/internal/worker/registry"
	"github.com/shadowapi/shadowapi/backend/pkg/query"
	"log/slog"
)

type SchedulerEmailJobArgs struct {
	SchedulerUUID string    `json:"scheduler_uuid"`
	JobUUID       string    `json:"job_uuid"`
	PipelineUUID  string    `json:"pipeline_uuid"`
	LastFetched   time.Time `json:"last_fetched"`
}

type MultiEmailScheduler struct {
	log        *slog.Logger
	dbp        *pgxpool.Pool
	queue      *queue.Queue
	cronParser cron.Parser
	interval   time.Duration
	maxBackoff time.Duration
	monitor    *monitor.WorkerMonitor
}

func NewMultiEmailScheduler(log *slog.Logger, dbp *pgxpool.Pool, queue *queue.Queue, monitor *monitor.WorkerMonitor) *MultiEmailScheduler {
	return &MultiEmailScheduler{
		log:        log,
		dbp:        dbp,
		queue:      queue,
		monitor:    monitor,
		cronParser: cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow),
		interval:   time.Minute,
		maxBackoff: 10 * time.Minute,
	}
}

func (s *MultiEmailScheduler) Start(ctx context.Context) {
	ticker := time.NewTicker(s.interval)
	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				s.run(ctx)
			case <-ctx.Done():
				s.log.Info("MultiEmailScheduler shutting down")
				return
			}
		}
	}()
}

func (s *MultiEmailScheduler) run(ctx context.Context) {
	queries := query.New(s.dbp)
	now := time.Now().UTC()

	// Get all enabled, unpaused email schedulers.
	schedulers, err := queries.GetSchedulers(ctx, query.GetSchedulersParams{
		OrderBy:        "created_at",
		OrderDirection: "asc",
		Offset:         0,
		Limit:          100,
		PipelineUuid:   "",
		IsEnabled:      1,
		IsPaused:       0,
	})
	if err != nil {
		s.log.Error("Failed fetching schedulers", "err", err)
		return
	}

	if len(schedulers) == 0 {
		s.log.Debug("No schedulers found")
		return
	}

	// Loop over each scheduler row.
	for _, sched := range schedulers {

		jobUUID := uuid.Must(uuid.NewV7()).String()
		// TODO @reactima consider to make in transaction

		// If NextRun is set and still in the future, skip this scheduler.
		if sched.NextRun.Valid && sched.NextRun.Time.After(now) {
			s.log.Debug("MultiEmailScheduler Skipping scheduler", "schedulerUUID", sched.UUID.String(), "nextRun", sched.NextRun.Time)

			// TODO @reactima remove dummy job
			jobArgs := SchedulerEmailJobArgs{
				SchedulerUUID: sched.UUID.String(),
				JobUUID:       jobUUID,
				PipelineUUID:  sched.PipelineUuid.String(),
				LastFetched:   now,
			}
			jobPayload, err := json.Marshal(jobArgs)
			if err != nil {
				s.log.Error("Failed to marshal dummy job payload", "scheduler", sched.UUID.String(), "err", err)
				continue
			}
			headers := queue.Headers{"X-Job-ID": jobUUID}

			err = s.queue.PublishWithHeaders(ctx, registry.WorkerSubjectDummy, headers, jobPayload)
			if err != nil {
				s.log.Error("Failed to publish dummy job", "schedulerUUID", sched.UUID.String(), "pipelineUUID", sched.PipelineUuid.String(), "err", err)
				continue
			}
			continue
		}

		s.log.Debug("MultiEmailScheduler Scheduling job", "start schedulerUUID", sched.UUID.String(), "jobUUID", jobUUID)

		jobArgs := SchedulerEmailJobArgs{
			SchedulerUUID: sched.UUID.String(),
			JobUUID:       jobUUID,
			PipelineUUID:  sched.PipelineUuid.String(),
			LastFetched:   now,
		}
		jobPayload, err := json.Marshal(jobArgs)
		if err != nil {
			s.log.Error("Failed to marshal job payload", "scheduler", sched.UUID.String(), "err", err)
			continue
		}

		// TODO @reactima
		// 1. Check if the pipeline is enabled and not paused, check if pipeline is email_oauth
		// 2. Consider to check if previous is not running, and decide what to do ... , research best practices first ????
		headers := queue.Headers{"X-Job-ID": jobUUID}

		err = s.queue.PublishWithHeaders(ctx, registry.WorkerSubjectEmailOAuthFetch, headers, jobPayload)
		if err != nil {
			s.log.Error("Failed to publish job", "schedulerUUID", sched.UUID.String(), "pipelineUUID", sched.PipelineUuid.String(), "err", err)
			backoffDelay := s.calculateBackoff(sched)
			s.updateNextRun(ctx, queries, converter.UuidToPgUUID(sched.UUID), now.Add(backoffDelay))
			continue
		}

		// Calculate the next run time.
		nextRun := s.nextRunTime(sched, now)
		// Update the scheduler record with the new run time.
		s.updateSchedulerRun(ctx, queries, converter.UuidToPgUUID(sched.UUID), now, nextRun)
		// Increase the scheduled jobs metric.
		metrics.JobScheduledTotal.WithLabelValues(sched.PipelineUuid.String(), "").Inc()
	}
}

func (s *MultiEmailScheduler) nextRunTime(sch query.GetSchedulersRow, now time.Time) time.Time {
	if sch.ScheduleType == "cron" {
		schedule, err := s.cronParser.Parse(sch.CronExpression.String)
		if err == nil {
			return schedule.Next(now)
		}
	}
	return now.Add(24 * time.Hour)
}

func (s *MultiEmailScheduler) updateSchedulerRun(ctx context.Context, queries *query.Queries, id pgtype.UUID, lastRun, nextRun time.Time) {
	err := queries.UpdateScheduler(ctx, query.UpdateSchedulerParams{
		CronExpression: pgtype.Text{String: "", Valid: false},
		RunAt:          pgtype.Timestamptz{Time: lastRun, Valid: true},
		Timezone:       "UTC",
		NextRun:        pgtype.Timestamptz{Time: nextRun, Valid: true},
		LastRun:        pgtype.Timestamptz{Time: lastRun, Valid: true},
		IsEnabled:      true,
		IsPaused:       false,
		UUID:           id,
	})
	if err != nil {
		s.log.Error("Failed to update scheduler run times", "error", err)
	}
}

func (s *MultiEmailScheduler) calculateBackoff(sch query.GetSchedulersRow) time.Duration {
	baseDelay := 5 * time.Minute
	backoff := baseDelay
	if backoff > s.maxBackoff {
		backoff = s.maxBackoff
	}
	return backoff
}

func (s *MultiEmailScheduler) updateNextRun(ctx context.Context, queries *query.Queries, id pgtype.UUID, nextRun time.Time) {
	err := queries.UpdateScheduler(ctx, query.UpdateSchedulerParams{
		CronExpression: pgtype.Text{String: "", Valid: false},
		RunAt:          pgtype.Timestamptz{Time: time.Now().UTC(), Valid: true},
		Timezone:       "UTC",
		NextRun:        pgtype.Timestamptz{Time: nextRun, Valid: true},
		LastRun:        pgtype.Timestamptz{Valid: false},
		IsEnabled:      true,
		IsPaused:       false,
		UUID:           id,
	})
	if err != nil {
		s.log.Error("Failed to update scheduler next run", "error", err)
	}
}
