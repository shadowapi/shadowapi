// Package email provides email sending functionality
package email

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"log/slog"
	"net/smtp"
	"strings"

	"github.com/samber/do/v2"

	"github.com/shadowapi/shadowapi/backend/internal/config"
)

// Service handles email sending
type Service struct {
	cfg *config.Config
	log *slog.Logger
}

// Provide creates a new email service for dependency injection
func Provide(i do.Injector) (*Service, error) {
	return &Service{
		cfg: do.MustInvoke[*config.Config](i),
		log: do.MustInvoke[*slog.Logger](i),
	}, nil
}

// InviteEmailParams contains data for invite email
type InviteEmailParams struct {
	ToEmail       string
	InviterName   string
	WorkspaceName string
	Role          string
	InviteLink    string
	ExpiresIn     string // e.g., "24 hours"
}

// PasswordResetEmailParams contains data for password reset email
type PasswordResetEmailParams struct {
	ToEmail   string
	UserName  string
	ResetLink string
	ExpiresIn string // e.g., "30 minutes"
}

// SendInviteEmail sends a workspace invite email
func (s *Service) SendInviteEmail(ctx context.Context, params InviteEmailParams) error {
	tmpl, err := template.New("invite").Parse(inviteEmailTemplate)
	if err != nil {
		return fmt.Errorf("parse template: %w", err)
	}

	var body bytes.Buffer
	if err := tmpl.Execute(&body, params); err != nil {
		return fmt.Errorf("execute template: %w", err)
	}

	subject := fmt.Sprintf("You've been invited to join %s on MeshPump", params.WorkspaceName)
	return s.send(params.ToEmail, subject, body.String())
}

// SendPasswordResetEmail sends a password reset email
func (s *Service) SendPasswordResetEmail(ctx context.Context, params PasswordResetEmailParams) error {
	tmpl, err := template.New("password_reset").Parse(passwordResetEmailTemplate)
	if err != nil {
		return fmt.Errorf("parse template: %w", err)
	}

	var body bytes.Buffer
	if err := tmpl.Execute(&body, params); err != nil {
		return fmt.Errorf("execute template: %w", err)
	}

	subject := "Reset your MeshPump password"
	return s.send(params.ToEmail, subject, body.String())
}

func (s *Service) send(to, subject, htmlBody string) error {
	from := fmt.Sprintf("%s <%s>", s.cfg.SMTP.FromName, s.cfg.SMTP.From)

	msg := strings.Builder{}
	msg.WriteString(fmt.Sprintf("From: %s\r\n", from))
	msg.WriteString(fmt.Sprintf("To: %s\r\n", to))
	msg.WriteString(fmt.Sprintf("Subject: %s\r\n", subject))
	msg.WriteString("MIME-Version: 1.0\r\n")
	msg.WriteString("Content-Type: text/html; charset=UTF-8\r\n")
	msg.WriteString("\r\n")
	msg.WriteString(htmlBody)

	addr := fmt.Sprintf("%s:%d", s.cfg.SMTP.Host, s.cfg.SMTP.Port)

	var auth smtp.Auth
	if s.cfg.SMTP.Username != "" {
		auth = smtp.PlainAuth("", s.cfg.SMTP.Username, s.cfg.SMTP.Password, s.cfg.SMTP.Host)
	}

	s.log.Info("sending email", "to", to, "subject", subject, "smtp_host", addr)

	if err := smtp.SendMail(addr, auth, s.cfg.SMTP.From, []string{to}, []byte(msg.String())); err != nil {
		s.log.Error("failed to send email", "error", err, "to", to)
		return fmt.Errorf("send mail: %w", err)
	}

	return nil
}

const inviteEmailTemplate = `<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>Workspace Invitation</title>
</head>
<body style="font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; max-width: 600px; margin: 0 auto; padding: 20px; background: #f9f9f9;">
    <div style="background: #ffffff; padding: 40px; border-radius: 8px; box-shadow: 0 2px 4px rgba(0,0,0,0.1);">
        <div style="text-align: center; margin-bottom: 30px;">
            <h1 style="color: #1a1a1a; margin: 0; font-size: 24px;">You're invited to join {{.WorkspaceName}}</h1>
        </div>

        <p style="color: #4a4a4a; font-size: 16px; line-height: 1.6; margin-bottom: 20px;">
            {{.InviterName}} has invited you to join the <strong>{{.WorkspaceName}}</strong> workspace on MeshPump as a <strong>{{.Role}}</strong>.
        </p>

        <div style="text-align: center; margin: 30px 0;">
            <a href="{{.InviteLink}}" style="background: #1890ff; color: white; padding: 14px 32px; text-decoration: none; border-radius: 6px; font-weight: 500; display: inline-block; font-size: 16px;">
                Accept Invitation
            </a>
        </div>

        <p style="color: #666; font-size: 14px; line-height: 1.5;">
            This invitation will expire in <strong>{{.ExpiresIn}}</strong>. If you didn't expect this invitation, you can safely ignore this email.
        </p>

        <hr style="border: none; border-top: 1px solid #eee; margin: 30px 0;">

        <p style="color: #999; font-size: 12px; margin: 0;">
            If the button doesn't work, copy and paste this link into your browser:<br>
            <a href="{{.InviteLink}}" style="color: #1890ff; word-break: break-all;">{{.InviteLink}}</a>
        </p>
    </div>

    <p style="text-align: center; color: #999; font-size: 12px; margin-top: 20px;">
        MeshPump - Unified Messaging Platform
    </p>
</body>
</html>`

const passwordResetEmailTemplate = `<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>Password Reset</title>
</head>
<body style="font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; max-width: 600px; margin: 0 auto; padding: 20px; background: #f9f9f9;">
    <div style="background: #ffffff; padding: 40px; border-radius: 8px; box-shadow: 0 2px 4px rgba(0,0,0,0.1);">
        <div style="text-align: center; margin-bottom: 30px;">
            <h1 style="color: #1a1a1a; margin: 0; font-size: 24px;">Reset Your Password</h1>
        </div>

        <p style="color: #4a4a4a; font-size: 16px; line-height: 1.6; margin-bottom: 20px;">
            Hi{{if .UserName}} {{.UserName}}{{end}},
        </p>

        <p style="color: #4a4a4a; font-size: 16px; line-height: 1.6; margin-bottom: 20px;">
            We received a request to reset your password for your MeshPump account. Click the button below to create a new password.
        </p>

        <div style="text-align: center; margin: 30px 0;">
            <a href="{{.ResetLink}}" style="background: #1890ff; color: white; padding: 14px 32px; text-decoration: none; border-radius: 6px; font-weight: 500; display: inline-block; font-size: 16px;">
                Reset Password
            </a>
        </div>

        <p style="color: #666; font-size: 14px; line-height: 1.5;">
            This link will expire in <strong>{{.ExpiresIn}}</strong>. If you didn't request a password reset, you can safely ignore this email - your password will remain unchanged.
        </p>

        <hr style="border: none; border-top: 1px solid #eee; margin: 30px 0;">

        <p style="color: #999; font-size: 12px; margin: 0;">
            If the button doesn't work, copy and paste this link into your browser:<br>
            <a href="{{.ResetLink}}" style="color: #1890ff; word-break: break-all;">{{.ResetLink}}</a>
        </p>
    </div>

    <p style="text-align: center; color: #999; font-size: 12px; margin-top: 20px;">
        MeshPump - Unified Messaging Platform
    </p>
</body>
</html>`
