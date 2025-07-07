package zitadel

import (
	"log/slog"

	"golang.org/x/oauth2"
)

// LogExchangeError writes the full body of an *oauth2.RetrieveError.
func LogExchangeError(l *slog.Logger, err error) {
	if rErr, ok := err.(*oauth2.RetrieveError); ok {
		l.Error("token exchange failed",
			"status", rErr.Response.StatusCode,
			"body", string(rErr.Body))
		return
	}
	l.Error("token exchange failed", "error", err)
}
