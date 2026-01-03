// Package subjects defines NATS subject patterns for the distributed worker system
package subjects

import (
	"fmt"
)

const (
	// Prefix is the base prefix for all subjects
	Prefix = "shadowapi"

	// StreamName is the NATS JetStream stream for jobs
	StreamName = "jobs"

	// ResultsStreamName is the NATS JetStream stream for results
	ResultsStreamName = "results"
)

// Job types
const (
	JobTypeEmailOAuthFetch    = "emailOAuthFetch"
	JobTypeEmailApplyPipeline = "emailApplyPipeline"
	JobTypeTokenRefresh       = "tokenRefresh"
	JobTypeDummy              = "dummy"

	// Test connection job types
	JobTypeTestConnectionEmailOAuth = "testConnectionEmailOAuth"
	JobTypeTestConnectionPostgres   = "testConnectionPostgres"
)

// JobSubject returns the subject for publishing a job
// Format: shadowapi.jobs.workspace.{slug}.{jobType}
func JobSubject(workspaceSlug, jobType string) string {
	return fmt.Sprintf("%s.jobs.workspace.%s.%s", Prefix, workspaceSlug, jobType)
}

// GlobalJobSubject returns the subject for publishing a global job
// Format: shadowapi.jobs.global.{jobType}
func GlobalJobSubject(jobType string) string {
	return fmt.Sprintf("%s.jobs.global.%s", Prefix, jobType)
}

// AllJobsPattern returns the subject pattern to match all jobs
// Format: shadowapi.jobs.>
func AllJobsPattern() string {
	return fmt.Sprintf("%s.jobs.>", Prefix)
}

// WorkspaceJobsPattern returns the subject pattern to match all workspace jobs
// Format: shadowapi.jobs.workspace.{slug}.>
func WorkspaceJobsPattern(workspaceSlug string) string {
	return fmt.Sprintf("%s.jobs.workspace.%s.>", Prefix, workspaceSlug)
}

// ResultSubject returns the subject for publishing a job result
// Format: shadowapi.results.{jobID}
func ResultSubject(jobID string) string {
	return fmt.Sprintf("%s.results.%s", Prefix, jobID)
}

// AllResultsPattern returns the subject pattern to match all results
// Format: shadowapi.results.>
func AllResultsPattern() string {
	return fmt.Sprintf("%s.results.>", Prefix)
}

// AllJobSubjects returns all known job subject patterns for stream setup
func AllJobSubjects() []string {
	return []string{
		fmt.Sprintf("%s.jobs.>", Prefix),
	}
}

// AllResultSubjects returns all known result subject patterns for stream setup
func AllResultSubjects() []string {
	return []string{
		fmt.Sprintf("%s.results.>", Prefix),
	}
}
