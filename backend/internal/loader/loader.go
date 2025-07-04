package loader

import (
	"context"
	"log/slog"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/samber/do/v2"
	"golang.org/x/crypto/bcrypt"

	"github.com/shadowapi/shadowapi/backend/internal/config"
)

type Loader struct {
	cfg *config.Config
	log *slog.Logger
	dbp *pgxpool.Pool
}

// Provide loader instance for the dependency injector
func Provide(i do.Injector) (*Loader, error) {
	log := do.MustInvoke[*slog.Logger](i)
	log.Debug("Registering loader")
	return &Loader{
		cfg: do.MustInvoke[*config.Config](i),
		log: log,
		dbp: do.MustInvoke[*pgxpool.Pool](i),
	}, nil
}

// Run starts the loader
func (s *Loader) Run(ctx context.Context) error {
	if s.cfg.InitAdmin.Email != "" && s.cfg.InitAdmin.Password != "" {
		var count int
		if err := s.dbp.QueryRow(ctx, `SELECT COUNT(*) FROM "user"`).Scan(&count); err != nil {
			return err
		}
		if count == 0 {
			hash, err := bcrypt.GenerateFromPassword([]byte(s.cfg.InitAdmin.Password), bcrypt.DefaultCost)
			if err != nil {
				return err
			}
			_, err = s.dbp.Exec(ctx, `INSERT INTO "user" (uuid,email,password,first_name,last_name,is_enabled,is_admin) VALUES (gen_random_uuid(),$1,$2,'Admin','User',true,true)`, s.cfg.InitAdmin.Email, string(hash))
			if err != nil {
				return err
			}
			s.log.Info("created initial admin user", "email", s.cfg.InitAdmin.Email)
		}
	}
	return nil
}
