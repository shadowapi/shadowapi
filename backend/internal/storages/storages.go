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
// row's settings into api.StoragePostgres and builds a new *pgxpool.Pool.
// The resulting map is keyed by the storage.UUID string.
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

// TableInfo holds basic information about a database table.
type TableInfo struct {
	Name          string
	RowCount      int64
	HasPrimaryKey bool
}

// TableSchema holds the schema information for a table.
type TableSchema struct {
	Name     string
	Exists   bool
	Fields   []FieldInfo
	RowCount int64
}

// FieldInfo holds information about a table column.
type FieldInfo struct {
	Name         string
	Type         string // Our type (TEXT, INTEGER, etc.)
	PgType       string // Original PostgreSQL type
	Nullable     bool
	IsPrimaryKey bool
	DefaultValue string
}

// Manager provides storage operations including introspection.
type Manager struct {
	log *slog.Logger
	dbp *pgxpool.Pool
}

// NewManager creates a new storage manager.
func NewManager(log *slog.Logger, dbp *pgxpool.Pool) *Manager {
	return &Manager{log: log, dbp: dbp}
}

// ProvideManager provides the storage manager via DI.
func ProvideManager(i do.Injector) (*Manager, error) {
	log := do.MustInvoke[*slog.Logger](i)
	dbp := do.MustInvoke[*pgxpool.Pool](i)
	return NewManager(log, dbp), nil
}

// GetConnectionPool creates a temporary connection pool to the specified storage's database.
func (m *Manager) GetConnectionPool(ctx context.Context, storageUUID string) (*pgxpool.Pool, error) {
	storage, err := m.getStorageSettings(ctx, storageUUID)
	if err != nil {
		return nil, err
	}

	user := storage.User.Or("")
	pass := storage.Password.Or("")
	host := storage.Host.Or("")
	port := storage.Port.Or("")
	database := storage.Database.Or("postgres")
	options := storage.Options.Or("")

	if user == "" || host == "" {
		return nil, fmt.Errorf("missing required connection fields")
	}

	uri := buildPostgresURIWithDB(user, pass, host, port, database, options)
	cfg, err := pgxpool.ParseConfig(uri)
	if err != nil {
		return nil, fmt.Errorf("failed to parse connection URI: %w", err)
	}

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return pool, nil
}

// ListTables returns all tables in the public schema of the connected database.
func (m *Manager) ListTables(ctx context.Context, storageUUID string) ([]TableInfo, error) {
	pool, err := m.GetConnectionPool(ctx, storageUUID)
	if err != nil {
		return nil, err
	}
	defer pool.Close()

	query := `
		SELECT
			t.table_name,
			COALESCE((SELECT reltuples::bigint FROM pg_class WHERE relname = t.table_name AND relnamespace = 'public'::regnamespace), 0) as row_count,
			EXISTS (
				SELECT 1 FROM information_schema.table_constraints tc
				WHERE tc.table_schema = 'public'
				  AND tc.table_name = t.table_name
				  AND tc.constraint_type = 'PRIMARY KEY'
			) as has_primary_key
		FROM information_schema.tables t
		WHERE t.table_schema = 'public'
		  AND t.table_type = 'BASE TABLE'
		ORDER BY t.table_name
	`

	rows, err := pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query tables: %w", err)
	}
	defer rows.Close()

	var tables []TableInfo
	for rows.Next() {
		var t TableInfo
		if err := rows.Scan(&t.Name, &t.RowCount, &t.HasPrimaryKey); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		tables = append(tables, t)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return tables, nil
}

// GetTableFields returns the schema information for a specific table.
func (m *Manager) GetTableFields(ctx context.Context, storageUUID, tableName string) (*TableSchema, error) {
	pool, err := m.GetConnectionPool(ctx, storageUUID)
	if err != nil {
		return nil, err
	}
	defer pool.Close()

	// Check if table exists
	var exists bool
	existsQuery := `
		SELECT EXISTS (
			SELECT 1 FROM information_schema.tables
			WHERE table_schema = 'public' AND table_name = $1
		)
	`
	if err := pool.QueryRow(ctx, existsQuery, tableName).Scan(&exists); err != nil {
		return nil, fmt.Errorf("failed to check table existence: %w", err)
	}

	schema := &TableSchema{
		Name:   tableName,
		Exists: exists,
	}

	if !exists {
		return schema, nil
	}

	// Get row count
	rowCountQuery := `SELECT COALESCE(reltuples::bigint, 0) FROM pg_class WHERE relname = $1 AND relnamespace = 'public'::regnamespace`
	_ = pool.QueryRow(ctx, rowCountQuery, tableName).Scan(&schema.RowCount) // Ignore error, row count is optional

	// Get column information
	fieldsQuery := `
		SELECT
			c.column_name,
			c.data_type,
			c.is_nullable = 'YES',
			COALESCE(
				EXISTS (
					SELECT 1
					FROM information_schema.table_constraints tc
					JOIN information_schema.key_column_usage kcu
						ON tc.constraint_name = kcu.constraint_name
						AND tc.table_schema = kcu.table_schema
					WHERE tc.table_schema = 'public'
					  AND tc.table_name = $1
					  AND tc.constraint_type = 'PRIMARY KEY'
					  AND kcu.column_name = c.column_name
				),
				false
			) as is_primary_key,
			COALESCE(c.column_default, '') as default_value
		FROM information_schema.columns c
		WHERE c.table_schema = 'public'
		  AND c.table_name = $1
		ORDER BY c.ordinal_position
	`

	rows, err := pool.Query(ctx, fieldsQuery, tableName)
	if err != nil {
		return nil, fmt.Errorf("failed to query columns: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var f FieldInfo
		if err := rows.Scan(&f.Name, &f.PgType, &f.Nullable, &f.IsPrimaryKey, &f.DefaultValue); err != nil {
			return nil, fmt.Errorf("failed to scan column: %w", err)
		}
		f.Type = mapPgTypeToFieldType(f.PgType)
		schema.Fields = append(schema.Fields, f)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating columns: %w", err)
	}

	return schema, nil
}

// CreateTable creates a new table in the database.
func (m *Manager) CreateTable(ctx context.Context, storageUUID, tableName string, fields []api.StoragePostgresField, dropIfExists bool) (wasDropped bool, err error) {
	pool, err := m.GetConnectionPool(ctx, storageUUID)
	if err != nil {
		return false, err
	}
	defer pool.Close()

	// Start transaction
	tx, err := pool.Begin(ctx)
	if err != nil {
		return false, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Check if table exists
	var exists bool
	existsQuery := `SELECT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_schema = 'public' AND table_name = $1)`
	if err := tx.QueryRow(ctx, existsQuery, tableName).Scan(&exists); err != nil {
		return false, fmt.Errorf("failed to check table existence: %w", err)
	}

	if exists {
		if !dropIfExists {
			return false, fmt.Errorf("table %s already exists", tableName)
		}
		// Drop the existing table
		dropQuery := fmt.Sprintf("DROP TABLE IF EXISTS %s", quoteIdentifier(tableName))
		if _, err := tx.Exec(ctx, dropQuery); err != nil {
			return false, fmt.Errorf("failed to drop table: %w", err)
		}
		wasDropped = true
	}

	// Build CREATE TABLE statement
	createSQL := buildCreateTableSQL(tableName, fields)
	if _, err := tx.Exec(ctx, createSQL); err != nil {
		return wasDropped, fmt.Errorf("failed to create table: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return wasDropped, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return wasDropped, nil
}

// TableExists checks if a table exists in the database.
func (m *Manager) TableExists(ctx context.Context, storageUUID, tableName string) (bool, error) {
	pool, err := m.GetConnectionPool(ctx, storageUUID)
	if err != nil {
		return false, err
	}
	defer pool.Close()

	var exists bool
	query := `SELECT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_schema = 'public' AND table_name = $1)`
	if err := pool.QueryRow(ctx, query, tableName).Scan(&exists); err != nil {
		return false, fmt.Errorf("failed to check table existence: %w", err)
	}

	return exists, nil
}

func (m *Manager) getStorageSettings(ctx context.Context, storageUUID string) (*api.StoragePostgres, error) {
	id, err := ConvertStringToPgUUID(storageUUID)
	if err != nil {
		return nil, fmt.Errorf("invalid storage UUID: %w", err)
	}

	row, err := query.New(m.dbp).GetStorage(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get storage: %w", err)
	}

	if row.Storage.Type != "postgres" {
		return nil, fmt.Errorf("storage is not of type postgres")
	}

	var settings api.StoragePostgres
	if err := json.Unmarshal(row.Storage.Settings, &settings); err != nil {
		return nil, fmt.Errorf("failed to unmarshal settings: %w", err)
	}

	return &settings, nil
}

// buildPostgresURIWithDB builds a PostgreSQL URI with database name.
func buildPostgresURIWithDB(user, pass, host, port, database, opts string) string {
	if database == "" {
		database = "postgres"
	}
	if port == "" {
		port = "5432"
	}
	if opts == "" {
		return fmt.Sprintf("postgres://%s:%s@%s:%s/%s", user, pass, host, port, database)
	}
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?%s", user, pass, host, port, database, opts)
}

// mapPgTypeToFieldType maps PostgreSQL data types to our field types.
func mapPgTypeToFieldType(pgType string) string {
	switch pgType {
	case "text", "character varying", "character", "varchar", "char", "name", "uuid":
		return "TEXT"
	case "integer", "smallint", "bigint", "serial", "smallserial", "bigserial", "int2", "int4", "int8":
		return "INTEGER"
	case "boolean", "bool":
		return "BOOLEAN"
	case "timestamp without time zone", "timestamp with time zone", "timestamp", "timestamptz",
		"date", "time without time zone", "time with time zone", "time", "timetz":
		return "TIMESTAMP"
	case "json", "jsonb":
		return "JSONB"
	default:
		return "TEXT" // Default to TEXT for unknown types
	}
}

// quoteIdentifier quotes a PostgreSQL identifier.
func quoteIdentifier(name string) string {
	return fmt.Sprintf("\"%s\"", name)
}

// buildCreateTableSQL builds a CREATE TABLE statement from field definitions.
func buildCreateTableSQL(tableName string, fields []api.StoragePostgresField) string {
	var columns []string
	var primaryKey string

	for _, f := range fields {
		col := fmt.Sprintf("%s %s", quoteIdentifier(f.Name), fieldTypeToSQL(string(f.Type)))

		// Handle nullable
		nullable := true
		if f.Nullable.IsSet() {
			nullable = f.Nullable.Value
		}
		if !nullable {
			col += " NOT NULL"
		}

		// Handle default value
		if f.DefaultValue.IsSet() && f.DefaultValue.Value != "" {
			col += fmt.Sprintf(" DEFAULT %s", f.DefaultValue.Value)
		}

		columns = append(columns, col)

		// Track primary key
		if f.IsPrimaryKey.IsSet() && f.IsPrimaryKey.Value {
			primaryKey = f.Name
		}
	}

	// Add primary key constraint
	if primaryKey != "" {
		columns = append(columns, fmt.Sprintf("PRIMARY KEY (%s)", quoteIdentifier(primaryKey)))
	}

	return fmt.Sprintf("CREATE TABLE %s (\n  %s\n)", quoteIdentifier(tableName), joinStrings(columns, ",\n  "))
}

// fieldTypeToSQL converts our field type to PostgreSQL data type.
func fieldTypeToSQL(fieldType string) string {
	switch fieldType {
	case "TEXT":
		return "TEXT"
	case "INTEGER":
		return "BIGINT"
	case "BOOLEAN":
		return "BOOLEAN"
	case "TIMESTAMP":
		return "TIMESTAMPTZ"
	case "JSONB":
		return "JSONB"
	default:
		return "TEXT"
	}
}

// joinStrings joins strings with a separator.
func joinStrings(strs []string, sep string) string {
	result := ""
	for i, s := range strs {
		if i > 0 {
			result += sep
		}
		result += s
	}
	return result
}
