package tenant

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ogen-go/ogen/middleware"
	"github.com/samber/do/v2"

	"github.com/shadowapi/shadowapi/backend/internal/config"
	"github.com/shadowapi/shadowapi/backend/pkg/query"
)

// Middleware handles tenant extraction from subdomain and context injection
type Middleware struct {
	log        *slog.Logger
	dbp        *pgxpool.Pool
	baseDomain string
}

// Provide creates a new tenant middleware via DI
func Provide(i do.Injector) (*Middleware, error) {
	cfg := do.MustInvoke[*config.Config](i)
	log := do.MustInvoke[*slog.Logger](i)
	dbp := do.MustInvoke[*pgxpool.Pool](i)

	return &Middleware{
		log:        log,
		dbp:        dbp,
		baseDomain: cfg.Tenant.BaseDomain,
	}, nil
}

// OgenMiddleware is the ogen middleware for tenant extraction
func (m *Middleware) OgenMiddleware(req middleware.Request, next middleware.Next) (middleware.Response, error) {
	subdomain := ExtractSubdomain(req.Raw.Host, m.baseDomain)

	// Allow requests to root domain (tenant selection page, public endpoints)
	if subdomain == "" {
		return next(req)
	}

	// Look up tenant by subdomain
	t, err := query.New(m.dbp).GetTenantByName(req.Context, subdomain)
	if err != nil {
		if err == pgx.ErrNoRows {
			m.log.Debug("tenant not found", "subdomain", subdomain)
			return middleware.Response{}, &errWithCode{
				status: http.StatusNotFound,
				err:    fmt.Errorf("tenant not found: %s", subdomain),
			}
		}
		m.log.Error("failed to get tenant", "subdomain", subdomain, "error", err)
		return middleware.Response{}, &errWithCode{
			status: http.StatusInternalServerError,
			err:    fmt.Errorf("failed to lookup tenant"),
		}
	}

	if !t.IsEnabled {
		m.log.Debug("tenant disabled", "subdomain", subdomain)
		return middleware.Response{}, &errWithCode{
			status: http.StatusForbidden,
			err:    fmt.Errorf("tenant disabled"),
		}
	}

	// Add tenant to context
	tenant := &Tenant{
		UUID:        t.UUID.String(),
		Name:        t.Name,
		DisplayName: t.DisplayName,
		IsEnabled:   t.IsEnabled,
	}
	ctx := WithTenant(req.Context, tenant)
	req.SetContext(ctx)

	m.log.Debug("tenant context set", "tenant", tenant.Name, "uuid", tenant.UUID)

	return next(req)
}

type errWithCode struct {
	err    error
	status int
}

func (e *errWithCode) Error() string   { return e.err.Error() }
func (e *errWithCode) StatusCode() int { return e.status }
