// backend/internal/storages/storages.go
package storages

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"log/slog"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/samber/do/v2"

	"github.com/shadowapi/shadowapi/backend/pkg/api"
	"github.com/shadowapi/shadowapi/backend/pkg/query"
)

// ProvideDynamicPGConnections uses your main Postgres pool (dbp) to query for
// storages of type "postgres" (using query.GetStorages). It unmarshals each
// row's settings into api.StoragePostgres and, depending on `IsSameDatabase`,
// either reuses dbp or builds a new *pgxpool.Pool. The resulting map is keyed
// by the storage.UUID string.
func ProvideDynamicPGConnections(i do.Injector) (map[string]*pgxpool.Pool, error) {
	ctx := do.MustInvoke[context.Context](i)
	log := do.MustInvoke[*slog.Logger](i)
	dbp := do.MustInvoke[*pgxpool.Pool](i)

	storages, err := loadStoragesByType(ctx, dbp, "postgres")
	if err != nil {
		return nil, fmt.Errorf("failed to load 'postgres' storages: %w", err)
	}

	pgMap := make(map[string]*pgxpool.Pool)
	for _, row := range storages {
		record := row.Storage // query.Storage
		uuidStr := record.UUID.String()

		var pgSettings api.StoragePostgres
		if err := json.Unmarshal(record.Settings, &pgSettings); err != nil {
			log.Error("failed to unmarshal postgres settings",
				"error", err,
				"storageUUID", uuidStr,
			)
			continue
		}

		if pgSettings.IsSameDatabase.IsSet() && pgSettings.IsSameDatabase.Value {
			// Reuse the main *pgxpool.Pool
			pgMap[uuidStr] = dbp
			log.Info("added Postgres connection (reuse main dbp)",
				"storageUUID", uuidStr,
			)
			continue
		}

		// Build a connection URI from the provided fields
		user := pgSettings.User.Or("")
		pass := pgSettings.Password.Or("")
		host := pgSettings.Host.Or("")
		port := pgSettings.Port.Or("")
		options := pgSettings.Options.Or("")

		if user == "" || host == "" || port == "" {
			log.Error("missing required Postgres fields for dynamic connection",
				"user", user, "host", host, "port", port,
				"storageUUID", uuidStr,
			)
			continue
		}

		uri := buildPostgresURI(user, pass, host, port, options)
		cfg, err := pgxpool.ParseConfig(uri)
		if err != nil {
			log.Error("failed to parse Postgres URI",
				"error", err,
				"storageUUID", uuidStr,
				"uri", uri,
			)
			continue
		}

		newPool, err := pgxpool.NewWithConfig(ctx, cfg)
		if err != nil {
			log.Error("failed to create new Postgres pool",
				"error", err,
				"storageUUID", uuidStr,
				"uri", uri,
			)
			continue
		}
		if err := newPool.Ping(ctx); err != nil {
			log.Error("failed to ping new Postgres pool",
				"error", err,
				"storageUUID", uuidStr,
				"uri", uri,
			)
			newPool.Close()
			continue
		}

		pgMap[uuidStr] = newPool
		log.Info("added Postgres connection", "storageUUID", uuidStr)
	}

	if len(pgMap) == 0 {
		return nil, fmt.Errorf("no valid 'postgres' storages found")
	}
	return pgMap, nil
}

// ProvideDynamicS3Connections uses your main Postgres pool (dbp) to query for
// storages of type "s3". It unmarshals each row's settings into api.StorageS3,
// then creates an S3 client. The resulting map is keyed by the storage.UUID
// string.
func ProvideDynamicS3Connections(i do.Injector) (map[string]*s3.S3, error) {
	ctx := do.MustInvoke[context.Context](i)
	log := do.MustInvoke[*slog.Logger](i)
	dbp := do.MustInvoke[*pgxpool.Pool](i)

	storages, err := loadStoragesByType(ctx, dbp, "s3")
	if err != nil {
		return nil, fmt.Errorf("failed to load 's3' storages: %w", err)
	}

	s3Map := make(map[string]*s3.S3)
	for _, row := range storages {
		record := row.Storage
		uuidStr := record.UUID.String()

		var s3Settings api.StorageS3
		if err := json.Unmarshal(record.Settings, &s3Settings); err != nil {
			log.Error("failed to unmarshal s3 settings",
				"error", err,
				"storageUUID", uuidStr,
			)
			continue
		}

		region := s3Settings.Region
		if region == "" {
			log.Error("empty s3 region in settings",
				"storageUUID", uuidStr,
			)
			continue
		}

		sess, err := session.NewSession(&aws.Config{
			Region: aws.String(region),
			// If needed, set credentials or endpoint. For example:
			// Credentials: credentials.NewStaticCredentials(
			//		s3Settings.AccessKeyID, s3Settings.SecretAccessKey, ""),
			// Endpoint: aws.String("..."),
			// S3ForcePathStyle: aws.Bool(true), etc.
		})
		if err != nil {
			log.Error("failed to create AWS session",
				"error", err,
				"storageUUID", uuidStr,
				"region", region,
			)
			continue
		}

		client := s3.New(sess)
		s3Map[uuidStr] = client
		log.Info("added S3 connection", "storageUUID", uuidStr, "region", region)
	}

	if len(s3Map) == 0 {
		return nil, fmt.Errorf("no valid 's3' storages found")
	}
	return s3Map, nil
}

// ReconnectDynamicPGConnections runs a background goroutine every 5 minutes to
// ping each Postgres connection in pgMap. If Ping fails, it will re-fetch the
// storage row by UUID, parse the updated settings, and build a new pool in place.
func ReconnectDynamicPGConnections(ctx context.Context, pgMap map[string]*pgxpool.Pool, i do.Injector) {
	log := do.MustInvoke[*slog.Logger](i)
	dbp := do.MustInvoke[*pgxpool.Pool](i)

	ticker := time.NewTicker(5 * time.Minute)
	go func() {
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				for uuidStr, pool := range pgMap {
					if err := pool.Ping(ctx); err != nil {
						log.Warn("postgres connection lost; reconnecting",
							"storageUUID", uuidStr, "error", err,
						)
						pool.Close()

						record, err := loadSingleStorageByUUID(ctx, log, dbp, uuidStr)
						if err != nil {
							log.Error("failed to fetch storage by UUID",
								"storageUUID", uuidStr, "error", err,
							)
							continue
						}

						var pgSettings api.StoragePostgres
						if err := json.Unmarshal(record.Settings, &pgSettings); err != nil {
							log.Error("failed to unmarshal Postgres settings for reconnect",
								"storageUUID", uuidStr, "error", err,
							)
							continue
						}

						if pgSettings.IsSameDatabase.IsSet() && pgSettings.IsSameDatabase.Value {
							pgMap[uuidStr] = dbp
							log.Info("reconnected Postgres with main dbp",
								"storageUUID", uuidStr,
							)
							continue
						}

						user := pgSettings.User.Or("")
						pass := pgSettings.Password.Or("")
						host := pgSettings.Host.Or("")
						port := pgSettings.Port.Or("")
						opts := pgSettings.Options.Or("")

						if user == "" || host == "" || port == "" {
							log.Error("missing Postgres fields; cannot reconnect",
								"storageUUID", uuidStr,
								"user", user, "host", host, "port", port,
							)
							continue
						}

						uri := buildPostgresURI(user, pass, host, port, opts)
						cfg, err := pgxpool.ParseConfig(uri)
						if err != nil {
							log.Error("failed to parse Postgres URI",
								"error", err,
								"storageUUID", uuidStr,
								"uri", uri,
							)
							continue
						}

						newPool, err := pgxpool.NewWithConfig(ctx, cfg)
						if err != nil {
							log.Error("failed to create new Postgres pool",
								"error", err,
								"storageUUID", uuidStr,
								"uri", uri,
							)
							continue
						}
						if err := newPool.Ping(ctx); err != nil {
							log.Error("failed to ping new Postgres pool",
								"error", err,
								"storageUUID", uuidStr,
								"uri", uri,
							)
							newPool.Close()
							continue
						}

						pgMap[uuidStr] = newPool
						log.Info("reconnected Postgres storage", "storageUUID", uuidStr)
					}
				}
			}
		}
	}()
}

// loadStoragesByType calls query.GetStorages with Type=the given type. This is
// adapted from your existing handlers.
func loadStoragesByType(ctx context.Context, dbp *pgxpool.Pool, storageType string) ([]query.GetStorageRow, error) {
	arg := query.GetStoragesParams{
		OrderBy:        "created_at",
		OrderDirection: "desc",
		Offset:         0,
		Limit:          10000,
		Type:           storageType,
		UUID:           pgtype.UUID{},
		IsEnabled:      -1, // -1 => ignore
		Name:           "",
	}
	rows, err := query.New(dbp).GetStorages(ctx, arg)
	if err != nil {
		return nil, err
	}
	return convertGetStoragesRows(rows), nil
}

// loadSingleStorageByUUID calls query.GetStorage with a specific UUID.
func loadSingleStorageByUUID(ctx context.Context, log *slog.Logger, dbp *pgxpool.Pool, uuidStr string) (query.Storage, error) {

	id, err := ConvertStringToPgUUID(uuidStr)
	if err != nil {
		log.Error("failed to convert UUID string to pgtype.UUID", "error", err, "uuidStr", uuidStr)
	}

	q := query.New(dbp)
	row, err := q.GetStorage(ctx, id)
	if err != nil {
		return query.Storage{}, err
	}
	return row.Storage, nil
}

func convertGetStoragesRows(rows []query.GetStoragesRow) []query.GetStorageRow {
	out := make([]query.GetStorageRow, 0, len(rows))
	for _, r := range rows {
		out = append(out, query.GetStorageRow{
			Storage: query.Storage{
				UUID:      r.UUID,
				Name:      r.Name,
				Type:      r.Type,
				IsEnabled: r.IsEnabled,
				Settings:  r.Settings,
				CreatedAt: r.CreatedAt,
				UpdatedAt: r.UpdatedAt,
			},
		})
	}
	return out
}

// buildPostgresURI composes a basic Postgres DSN from parts. You can customize
// this as needed.
func buildPostgresURI(user, pass, host, port, opts string) string {
	if opts == "" {
		return fmt.Sprintf("postgres://%s:%s@%s:%s/postgres", user, pass, host, port)
	}
	return fmt.Sprintf("postgres://%s:%s@%s:%s/postgres?%s", user, pass, host, port, opts)
}

func uuidToPgxUUID(s string) [16]byte {
	// parse the UUID string into 16 bytes. You can also use your existing logic
	// from the handler: `uToBytes(...)`.
	// We'll keep it straightforward here:
	out := [16]byte{}
	parsed, _ := fmt.Sscanf(s, "%08x-%04x-%04x-%04x-%012x",
		&out[0], &out[1], &out[2], &out[3], &out[4])
	_ = parsed // ignoring error, but ideally handle parse errors
	return out
}

func ConvertStringToPgUUID(u string) (pgtype.UUID, error) {
	parsed, err := uuid.FromString(u)
	if err != nil {
		return pgtype.UUID{}, err
	}
	var pgUUID pgtype.UUID
	copy(pgUUID.Bytes[:], parsed[:])
	pgUUID.Valid = true
	return pgUUID, nil
}
