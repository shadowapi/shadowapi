package handler

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"golang.org/x/crypto/bcrypt"

	"github.com/shadowapi/shadowapi/backend/internal/converter"
	"github.com/shadowapi/shadowapi/backend/internal/email"
	"github.com/shadowapi/shadowapi/backend/pkg/api"
	"github.com/shadowapi/shadowapi/backend/pkg/query"
)

const (
	passwordResetTokenLength = 32
	passwordResetExpiry      = 30 * time.Minute
	passwordResetCooldown    = 5 * time.Minute
)

// RequestPasswordReset initiates password reset flow.
//
// POST /password/reset
func (h *Handler) RequestPasswordReset(ctx context.Context, req *api.PasswordResetRequest) (api.RequestPasswordResetRes, error) {
	q := query.New(h.dbp)

	// Always return success to prevent email enumeration
	successResponse := &api.RequestPasswordResetOK{
		Message: api.NewOptString("If an account exists with this email, a reset link has been sent."),
	}

	// Check if user exists
	user, err := q.GetUserByEmail(ctx, req.Email)
	if err != nil {
		if err == pgx.ErrNoRows {
			// Don't reveal that email doesn't exist
			return successResponse, nil
		}
		h.log.Error("failed to get user", "error", err)
		return successResponse, nil
	}

	// Rate limiting: Check for recent reset request
	lastReset, err := q.GetLatestPasswordResetByEmail(ctx, req.Email)
	if err == nil {
		timeSinceLastRequest := time.Since(lastReset.CreatedAt.Time)
		if timeSinceLastRequest < passwordResetCooldown {
			waitTime := passwordResetCooldown - timeSinceLastRequest
			return nil, ErrWithCode(http.StatusTooManyRequests,
				E("Please wait %d minutes before requesting another reset", int(waitTime.Minutes())+1))
		}
	}

	// Generate secure token (reuse function from invite.go)
	token, err := generateSecureToken(passwordResetTokenLength)
	if err != nil {
		h.log.Error("failed to generate token", "error", err)
		return successResponse, nil
	}

	// Hash token for storage
	tokenHash := hashToken(token)

	// Create password reset record
	resetUUID := uuid.Must(uuid.NewV7())
	expiresAt := time.Now().Add(passwordResetExpiry)

	_, err = q.CreatePasswordReset(ctx, query.CreatePasswordResetParams{
		UUID:      resetUUID,
		UserUUID:  &user.UUID,
		Email:     req.Email,
		TokenHash: tokenHash,
		ExpiresAt: pgtype.Timestamptz{Time: expiresAt, Valid: true},
	})
	if err != nil {
		h.log.Error("failed to create password reset", "error", err)
		return successResponse, nil
	}

	// Build reset link
	resetLink := fmt.Sprintf("%s/reset-password/%s", h.cfg.CSRBaseURL, token)

	// Send email
	if err := h.emailService.SendPasswordResetEmail(ctx, email.PasswordResetEmailParams{
		ToEmail:   req.Email,
		UserName:  user.FirstName,
		ResetLink: resetLink,
		ExpiresIn: "30 minutes",
	}); err != nil {
		h.log.Error("failed to send password reset email", "error", err, "email", req.Email)
	}

	h.log.Info("password reset requested", "email", req.Email)

	return successResponse, nil
}

// GetPasswordResetByToken validates token for reset page.
//
// GET /password/reset/{token}
func (h *Handler) GetPasswordResetByToken(ctx context.Context, params api.GetPasswordResetByTokenParams) (api.GetPasswordResetByTokenRes, error) {
	q := query.New(h.dbp)

	tokenHash := hashToken(params.Token)
	reset, err := q.GetValidPasswordResetByTokenHash(ctx, tokenHash)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, ErrWithCode(http.StatusNotFound, E("Reset link not found or expired"))
		}
		h.log.Error("failed to get password reset", "error", err)
		return nil, ErrWithCode(http.StatusInternalServerError, E("Failed to validate reset link"))
	}

	// Mask email for privacy (show first 2 chars + domain)
	maskedEmail := maskEmail(reset.Email)

	return &api.PasswordResetInfo{
		Email:     maskedEmail,
		ExpiresAt: reset.ExpiresAt.Time,
	}, nil
}

// ConfirmPasswordReset sets new password.
//
// POST /password/reset/confirm
func (h *Handler) ConfirmPasswordReset(ctx context.Context, req *api.PasswordResetConfirm) (api.ConfirmPasswordResetRes, error) {
	q := query.New(h.dbp)

	// Find and validate reset token
	tokenHash := hashToken(req.Token)
	reset, err := q.GetValidPasswordResetByTokenHash(ctx, tokenHash)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, ErrWithCode(http.StatusNotFound, E("Reset link not found or expired"))
		}
		h.log.Error("failed to get password reset", "error", err)
		return nil, ErrWithCode(http.StatusInternalServerError, E("Failed to validate reset link"))
	}

	// Hash new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		h.log.Error("failed to hash password", "error", err)
		return nil, ErrWithCode(http.StatusInternalServerError, E("Failed to process password"))
	}

	// Update password
	userPgUUID := pgtype.UUID{Bytes: converter.GofrsToGoogleUUID(*reset.UserUUID), Valid: true}
	if err := q.UpdateUserPassword(ctx, query.UpdateUserPasswordParams{
		Password: string(hashedPassword),
		UUID:     userPgUUID,
	}); err != nil {
		h.log.Error("failed to update password", "error", err)
		return nil, ErrWithCode(http.StatusInternalServerError, E("Failed to update password"))
	}

	// Invalidate all reset tokens for this user
	if err := q.InvalidatePasswordResetsForUser(ctx, reset.UserUUID); err != nil {
		h.log.Warn("failed to invalidate reset tokens", "error", err)
	}

	h.log.Info("password reset completed", "user_uuid", reset.UserUUID)

	return &api.ConfirmPasswordResetOK{
		RedirectURL: api.NewOptString("/login"),
	}, nil
}

// maskEmail masks an email for privacy display
func maskEmail(emailAddr string) string {
	parts := strings.Split(emailAddr, "@")
	if len(parts) != 2 || len(parts[0]) < 2 {
		return "***@***"
	}
	return parts[0][:2] + "***@" + parts[1]
}
