package cmd

import (
	"context"
	"crypto/tls"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/samber/do/v2"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/shadowapi/shadowapi/backend/cmd/worker/internal/executor"
	"github.com/shadowapi/shadowapi/backend/cmd/worker/internal/workerconfig"
	workerv1 "github.com/shadowapi/shadowapi/backend/pkg/proto/worker/v1"
)

var connectCmd = &cobra.Command{
	Use:   "connect",
	Short: "Connect to backend and start processing jobs",
	Long: `Connect establishes a persistent connection to the ShadowAPI backend
and starts receiving jobs for processing.

Requires WORKER_ID and WORKER_SECRET environment variables to be set
(obtained from the 'enroll' command).`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		cfg := do.MustInvoke[*workerconfig.Config](injector)
		log := do.MustInvoke[*slog.Logger](injector)
		exec := do.MustInvoke[*executor.Executor](injector)

		if cfg.WorkerID == "" || cfg.WorkerSecret == "" {
			return fmt.Errorf("WORKER_ID and WORKER_SECRET must be set (run 'enroll' first)")
		}

		log.Info("connecting to backend",
			"server", cfg.Server,
			"worker_id", cfg.WorkerID[:8]+"...",
			"tls", cfg.TLS,
		)

		// Configure transport credentials based on TLS setting
		var creds credentials.TransportCredentials
		if cfg.TLS {
			creds = credentials.NewTLS(&tls.Config{})
		} else {
			creds = insecure.NewCredentials()
		}

		conn, err := grpc.NewClient(cfg.Server, grpc.WithTransportCredentials(creds))
		if err != nil {
			return fmt.Errorf("failed to connect: %w", err)
		}
		defer conn.Close()

		client := workerv1.NewWorkerServiceClient(conn)

		stream, err := client.Connect(ctx)
		if err != nil {
			return fmt.Errorf("failed to open stream: %w", err)
		}

		// Send authentication
		if err := stream.Send(&workerv1.WorkerMessage{
			Payload: &workerv1.WorkerMessage_Authenticate{
				Authenticate: &workerv1.Authenticate{
					WorkerId:     cfg.WorkerID,
					WorkerSecret: cfg.WorkerSecret,
				},
			},
		}); err != nil {
			return fmt.Errorf("failed to authenticate: %w", err)
		}

		// Receive auth ack
		resp, err := stream.Recv()
		if err != nil {
			return fmt.Errorf("failed to receive auth response: %w", err)
		}

		ack := resp.GetAuthenticateAck()
		if ack == nil || !ack.Success {
			errMsg := "unknown error"
			if ack != nil {
				errMsg = ack.ErrorMessage
			}
			return fmt.Errorf("authentication failed: %s", errMsg)
		}

		log.Info("connected to backend",
			"is_global", ack.IsGlobal,
			"workspaces", ack.AllowedWorkspaces,
		)

		// Setup signal handling
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

		// Mutex to protect stream writes from concurrent goroutines
		var streamMu sync.Mutex

		// Start heartbeat goroutine
		go func() {
			ticker := time.NewTicker(time.Duration(cfg.HeartbeatInterval) * time.Second)
			defer ticker.Stop()

			for {
				select {
				case <-ctx.Done():
					return
				case <-ticker.C:
					activeJobs := exec.ActiveJobs()
					status := workerv1.WorkerStatus_WORKER_STATUS_IDLE
					if activeJobs > 0 {
						status = workerv1.WorkerStatus_WORKER_STATUS_BUSY
					}

					streamMu.Lock()
					err := stream.Send(&workerv1.WorkerMessage{
						Payload: &workerv1.WorkerMessage_Heartbeat{
							Heartbeat: &workerv1.Heartbeat{
								Timestamp:  timestamppb.Now(),
								Status:     status,
								ActiveJobs: activeJobs,
								Capacity:   int32(cfg.Capacity),
							},
						},
					})
					streamMu.Unlock()

					if err != nil {
						log.Error("failed to send heartbeat", "error", err)
						cancel()
						return
					}
					log.Debug("heartbeat sent", "active_jobs", activeJobs, "status", status)
				}
			}
		}()

		// Message handling loop
		go func() {
			for {
				msg, err := stream.Recv()
				if err != nil {
					log.Error("stream error", "error", err)
					cancel()
					return
				}

				switch p := msg.Payload.(type) {
				case *workerv1.ServerMessage_HeartbeatAck:
					log.Debug("heartbeat acknowledged",
						"server_time", p.HeartbeatAck.ServerTime.AsTime(),
					)

				case *workerv1.ServerMessage_JobAssignment:
					job := p.JobAssignment
					log.Info("job assigned",
						"job_id", job.JobId,
						"job_type", job.JobType,
						"workspace", job.WorkspaceSlug,
					)

					// Execute job asynchronously
					go func() {
						// Create context with deadline if provided
						jobCtx := ctx
						if job.Deadline != nil {
							var jobCancel context.CancelFunc
							jobCtx, jobCancel = context.WithDeadline(ctx, job.Deadline.AsTime())
							defer jobCancel()
						}

						// Execute the job
						resultData, err := exec.Execute(jobCtx, job.JobType, job.JobData)

						// Build result message
						result := &workerv1.JobResult{
							JobId:      job.JobId,
							Success:    err == nil,
							ResultData: resultData,
						}
						if err != nil {
							result.ErrorMessage = err.Error()
						}

						// Send result back
						streamMu.Lock()
						sendErr := stream.Send(&workerv1.WorkerMessage{
							Payload: &workerv1.WorkerMessage_JobResult{
								JobResult: result,
							},
						})
						streamMu.Unlock()

						if sendErr != nil {
							log.Error("failed to send job result",
								"job_id", job.JobId,
								"error", sendErr,
							)
						} else {
							log.Info("job result sent",
								"job_id", job.JobId,
								"success", result.Success,
							)
						}
					}()

				case *workerv1.ServerMessage_Disconnect:
					log.Info("disconnect requested by server",
						"reason", p.Disconnect.Reason,
					)
					cancel()
					return
				}
			}
		}()

		// Wait for signal or context cancellation
		select {
		case <-ctx.Done():
			log.Info("context cancelled, shutting down")
		case sig := <-sigCh:
			log.Info("received signal, shutting down", "signal", sig)
			cancel()
		}

		return nil
	},
}

func init() {
	LoadWorkerConfig(connectCmd)
	rootCmd.AddCommand(connectCmd)
}
