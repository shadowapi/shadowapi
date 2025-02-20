// Code generated by ogen, DO NOT EDIT.

package api

import (
	"context"
	"net/http"
	"strings"

	"github.com/go-faster/errors"

	"github.com/ogen-go/ogen/ogenerrors"
)

// SecurityHandler is handler for security parameters.
type SecurityHandler interface {
	// HandleBearerAuth handles BearerAuth security.
	HandleBearerAuth(ctx context.Context, operationName OperationName, t BearerAuth) (context.Context, error)
	// HandleSessionCookieAuth handles SessionCookieAuth security.
	HandleSessionCookieAuth(ctx context.Context, operationName OperationName, t SessionCookieAuth) (context.Context, error)
}

func findAuthorization(h http.Header, prefix string) (string, bool) {
	v, ok := h["Authorization"]
	if !ok {
		return "", false
	}
	for _, vv := range v {
		scheme, value, ok := strings.Cut(vv, " ")
		if !ok || !strings.EqualFold(scheme, prefix) {
			continue
		}
		return value, true
	}
	return "", false
}

func (s *Server) securityBearerAuth(ctx context.Context, operationName OperationName, req *http.Request) (context.Context, bool, error) {
	var t BearerAuth
	token, ok := findAuthorization(req.Header, "Bearer")
	if !ok {
		return ctx, false, nil
	}
	t.Token = token
	rctx, err := s.sec.HandleBearerAuth(ctx, operationName, t)
	if errors.Is(err, ogenerrors.ErrSkipServerSecurity) {
		return nil, false, nil
	} else if err != nil {
		return nil, false, err
	}
	return rctx, true, err
}
func (s *Server) securitySessionCookieAuth(ctx context.Context, operationName OperationName, req *http.Request) (context.Context, bool, error) {
	var t SessionCookieAuth
	const parameterName = "ory_kratos_session"
	var value string
	switch cookie, err := req.Cookie(parameterName); {
	case err == nil: // if NO error
		value = cookie.Value
	case errors.Is(err, http.ErrNoCookie):
		return ctx, false, nil
	default:
		return nil, false, errors.Wrap(err, "get cookie value")
	}
	t.APIKey = value
	rctx, err := s.sec.HandleSessionCookieAuth(ctx, operationName, t)
	if errors.Is(err, ogenerrors.ErrSkipServerSecurity) {
		return nil, false, nil
	} else if err != nil {
		return nil, false, err
	}
	return rctx, true, err
}

// SecuritySource is provider of security values (tokens, passwords, etc.).
type SecuritySource interface {
	// BearerAuth provides BearerAuth security value.
	BearerAuth(ctx context.Context, operationName OperationName) (BearerAuth, error)
	// SessionCookieAuth provides SessionCookieAuth security value.
	SessionCookieAuth(ctx context.Context, operationName OperationName) (SessionCookieAuth, error)
}

func (s *Client) securityBearerAuth(ctx context.Context, operationName OperationName, req *http.Request) error {
	t, err := s.sec.BearerAuth(ctx, operationName)
	if err != nil {
		return errors.Wrap(err, "security source \"BearerAuth\"")
	}
	req.Header.Set("Authorization", "Bearer "+t.Token)
	return nil
}
func (s *Client) securitySessionCookieAuth(ctx context.Context, operationName OperationName, req *http.Request) error {
	t, err := s.sec.SessionCookieAuth(ctx, operationName)
	if err != nil {
		return errors.Wrap(err, "security source \"SessionCookieAuth\"")
	}
	req.AddCookie(&http.Cookie{
		Name:  "ory_kratos_session",
		Value: t.APIKey,
	})
	return nil
}
