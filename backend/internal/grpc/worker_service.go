package grpc

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"log/slog"

	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/shadowapi/shadowapi/backend/pkg/query"

	workerv1 "github.com/shadowapi/shadowapi/backend/pkg/proto/worker/v1"
)

// WorkerService implements the gRPC WorkerService
type WorkerService struct {
	workerv1.UnimplementedWorkerServiceServer
	log     *slog.Logger
	dbp     *pgxpool.Pool
	manager *WorkerManager
}

// NewWorkerService creates a new WorkerService
func NewWorkerService(log *slog.Logger, dbp *pgxpool.Pool, manager *WorkerManager) *WorkerService {
	return &WorkerService{
		log:     log.With("service", "worker-grpc"),
		dbp:     dbp,
		manager: manager,
	}
}

// Enroll exchanges a one-time enrollment token for permanent worker credentials
func (s *WorkerService) Enroll(ctx context.Context, req *workerv1.EnrollRequest) (*workerv1.EnrollResponse, error) {
	if req.EnrollmentToken == "" {
		return nil, status.Error(codes.InvalidArgument, "enrollment_token required")
	}
	if req.WorkerName == "" {
		return nil, status.Error(codes.InvalidArgument, "worker_name required")
	}

	q := query.New(s.dbp)

	// Hash the provided token to compare with stored hash
	tokenHash := hashToken(req.EnrollmentToken)

	// Find valid token
	token, err := q.GetValidEnrollmentTokenByHash(ctx, tokenHash)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			s.log.Warn("enrollment failed: invalid or expired token")
			return nil, status.Error(codes.Unauthenticated, "invalid or expired token")
		}
		s.log.Error("database error during enrollment", "error", err)
		return nil, status.Error(codes.Internal, "database error")
	}

	// Generate worker credentials
	workerUUID := uuid.Must(uuid.NewV7())
	workerSecret := generateSecret(32)
	secretHash, err := bcrypt.GenerateFromPassword([]byte(workerSecret), bcrypt.DefaultCost)
	if err != nil {
		s.log.Error("failed to hash secret", "error", err)
		return nil, status.Error(codes.Internal, "failed to hash secret")
	}

	// Create worker record
	_, err = q.CreateRegisteredWorker(ctx, query.CreateRegisteredWorkerParams{
		UUID:       workerUUID,
		Name:       req.WorkerName,
		SecretHash: string(secretHash),
		Status:     "offline",
		IsGlobal:   token.IsGlobal,
		Version:    pgtype.Text{String: req.WorkerVersion, Valid: req.WorkerVersion != ""},
		Labels:     marshalLabels(req.Labels),
	})
	if err != nil {
		s.log.Error("failed to create worker", "error", err)
		return nil, status.Error(codes.Internal, "failed to create worker")
	}

	// Assign workspaces from token
	for _, wsUUID := range token.WorkspaceUuids {
		if !wsUUID.Valid {
			continue
		}
		linkUUID := uuid.Must(uuid.NewV7())
		// Convert pgtype.UUID to gofrs uuid
		workspaceUUID, err := uuid.FromBytes(wsUUID.Bytes[:])
		if err != nil {
			s.log.Warn("invalid workspace UUID in token", "error", err)
			continue
		}
		err = q.AddWorkerWorkspace(ctx, query.AddWorkerWorkspaceParams{
			UUID:          linkUUID,
			WorkerUUID:    &workerUUID,
			WorkspaceUUID: &workspaceUUID,
		})
		if err != nil {
			s.log.Warn("failed to add worker workspace", "worker", workerUUID, "workspace", workspaceUUID, "error", err)
		}
	}

	// Mark token as used
	err = q.MarkTokenUsed(ctx, query.MarkTokenUsedParams{
		UUID:             token.UUID,
		UsedByWorkerUuid: &workerUUID,
	})
	if err != nil {
		s.log.Warn("failed to mark token as used", "error", err)
	}

	// Get workspace slugs
	workspaces, err := q.GetWorkerWorkspaces(ctx, &workerUUID)
	if err != nil {
		s.log.Warn("failed to get worker workspaces", "error", err)
	}
	slugs := make([]string, len(workspaces))
	for i, ws := range workspaces {
		slugs[i] = ws.Slug
	}

	s.log.Info("worker enrolled",
		"worker_id", workerUUID,
		"name", req.WorkerName,
		"is_global", token.IsGlobal,
		"workspaces", slugs,
	)

	return &workerv1.EnrollResponse{
		WorkerId:          workerUUID.String(),
		WorkerSecret:      workerSecret,
		AllowedWorkspaces: slugs,
		IsGlobal:          token.IsGlobal,
	}, nil
}

// Connect establishes a bidirectional stream for heartbeats and job dispatch
func (s *WorkerService) Connect(stream workerv1.WorkerService_ConnectServer) error {
	ctx := stream.Context()

	// Get client address
	var clientAddr string
	if p, ok := peer.FromContext(ctx); ok {
		clientAddr = p.Addr.String()
	}

	// First message must be Authenticate
	msg, err := stream.Recv()
	if err != nil {
		return err
	}

	auth := msg.GetAuthenticate()
	if auth == nil {
		return status.Error(codes.Unauthenticated, "first message must be authenticate")
	}

	// Validate credentials
	workerUUID, err := uuid.FromString(auth.WorkerId)
	if err != nil {
		return status.Error(codes.InvalidArgument, "invalid worker_id")
	}

	q := query.New(s.dbp)
	worker, err := q.GetRegisteredWorker(ctx, workerUUID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return status.Error(codes.Unauthenticated, "worker not found")
		}
		s.log.Error("database error during authentication", "error", err)
		return status.Error(codes.Internal, "database error")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(worker.SecretHash), []byte(auth.WorkerSecret)); err != nil {
		s.log.Warn("worker authentication failed", "worker_id", workerUUID, "from", clientAddr)
		return status.Error(codes.Unauthenticated, "invalid credentials")
	}

	// Mark worker connected
	err = q.UpdateWorkerConnected(ctx, query.UpdateWorkerConnectedParams{
		UUID:          workerUUID,
		ConnectedFrom: pgtype.Text{String: clientAddr, Valid: clientAddr != ""},
	})
	if err != nil {
		s.log.Warn("failed to update worker connected status", "error", err)
	}

	// Get allowed workspaces
	workspaces, err := q.GetWorkerWorkspaces(ctx, &workerUUID)
	if err != nil {
		s.log.Warn("failed to get worker workspaces", "error", err)
	}
	slugs := make([]string, len(workspaces))
	for i, ws := range workspaces {
		slugs[i] = ws.Slug
	}

	// Send auth ack
	if err := stream.Send(&workerv1.ServerMessage{
		Payload: &workerv1.ServerMessage_AuthenticateAck{
			AuthenticateAck: &workerv1.AuthenticateAck{
				Success:           true,
				AllowedWorkspaces: slugs,
				IsGlobal:          worker.IsGlobal,
			},
		},
	}); err != nil {
		return err
	}

	// Register in manager
	conn := s.manager.Register(workerUUID.String(), worker.Name, stream, worker.IsGlobal, slugs)
	defer func() {
		s.manager.Unregister(workerUUID.String())
		_ = q.UpdateWorkerDisconnected(ctx, workerUUID)
		s.log.Info("worker disconnected", "worker_id", workerUUID, "name", worker.Name)
	}()

	s.log.Info("worker connected",
		"worker_id", workerUUID,
		"name", worker.Name,
		"from", clientAddr,
		"is_global", worker.IsGlobal,
		"workspaces", slugs,
	)

	// Handle messages
	for {
		msg, err := stream.Recv()
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}

		switch p := msg.Payload.(type) {
		case *workerv1.WorkerMessage_Heartbeat:
			err = q.UpdateWorkerHeartbeat(ctx, query.UpdateWorkerHeartbeatParams{
				UUID:   workerUUID,
				Status: statusToString(p.Heartbeat.Status),
			})
			if err != nil {
				s.log.Warn("failed to update heartbeat", "error", err)
			}
			conn.UpdateStatus(p.Heartbeat)

			if err := stream.Send(&workerv1.ServerMessage{
				Payload: &workerv1.ServerMessage_HeartbeatAck{
					HeartbeatAck: &workerv1.HeartbeatAck{
						ServerTime: timestamppb.Now(),
					},
				},
			}); err != nil {
				return err
			}

		case *workerv1.WorkerMessage_JobResult:
			// Future: handle job results
			s.log.Debug("job result received",
				"worker_id", workerUUID,
				"job_id", p.JobResult.JobId,
				"success", p.JobResult.Success,
			)
		}
	}
}

// hashToken creates a SHA256 hash of the token for storage comparison
func hashToken(token string) string {
	h := sha256.New()
	h.Write([]byte(token))
	return hex.EncodeToString(h.Sum(nil))
}

// generateSecret generates a cryptographically secure random secret
func generateSecret(length int) string {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		panic(err) // Should never happen
	}
	return hex.EncodeToString(bytes)
}

// marshalLabels converts labels map to JSON bytes
func marshalLabels(labels map[string]string) []byte {
	if len(labels) == 0 {
		return []byte("{}")
	}
	// Simple JSON encoding for labels
	result := "{"
	first := true
	for k, v := range labels {
		if !first {
			result += ","
		}
		result += `"` + k + `":"` + v + `"`
		first = false
	}
	result += "}"
	return []byte(result)
}

// statusToString converts WorkerStatus enum to string for database storage
func statusToString(s workerv1.WorkerStatus) string {
	switch s {
	case workerv1.WorkerStatus_WORKER_STATUS_IDLE:
		return "online"
	case workerv1.WorkerStatus_WORKER_STATUS_BUSY:
		return "online"
	case workerv1.WorkerStatus_WORKER_STATUS_DRAINING:
		return "draining"
	default:
		return "offline"
	}
}
