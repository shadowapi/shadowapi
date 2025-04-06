// Sqlc usage requires lot of type conversions, so we put them in a separate package
package converter

import (
	"errors"
	"fmt"
	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/shadowapi/shadowapi/backend/pkg/api"
	"strings"
)

func BytesToUUID(b []byte) ([16]byte, error) {
	var arr [16]byte
	if len(b) != 16 {
		return arr, errors.New("invalid UUID length")
	}
	copy(arr[:], b)
	return arr, nil
}

func UuidPtrToPgUUID(u *uuid.UUID) pgtype.UUID {
	if u == nil {
		return pgtype.UUID{Valid: false}
	}
	var arr [16]byte
	copy(arr[:], u.Bytes())
	return pgtype.UUID{Bytes: arr, Valid: true}
}

// Helper function to convert gofrs/uuid.UUID to pgx/pgtype.UUID.
func UuidToPgUUID(u uuid.UUID) pgtype.UUID {
	var pg pgtype.UUID
	copy(pg.Bytes[:], u.Bytes())
	pg.Valid = true
	return pg
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

// ConvertUUID converts a non-empty UUID string to a pointer to uuid.UUID.
// If the input is empty or invalid, it returns nil.
func ConvertUUID(originalUUID string) *uuid.UUID {
	if originalUUID == "" {
		return nil
	}
	u, err := uuid.FromString(originalUUID)
	if err != nil {
		return nil
	}
	return &u
}

// ConvertOptStringToUUID converts an api.OptString to a uuid.UUID.
// It returns an error if the OptString is not set or if the inner value is invalid.
func ConvertOptStringToUUID(opt api.OptString) (uuid.UUID, error) {
	if !opt.IsSet() || opt.Value == "" {
		return uuid.Nil, fmt.Errorf("opt string is not set")
	}
	return uuid.FromString(opt.Value)
}

func UToBytes(u uuid.UUID) [16]byte {
	var arr [16]byte
	copy(arr[:], u.Bytes())
	return arr
}

// --- Helper conversion functions ---

func ConvertOptNilStringToPgText(o api.OptNilString) pgtype.Text {
	if !o.IsSet() || o.IsNull() {
		return pgtype.Text{Valid: false}
	}
	return pgtype.Text{String: o.Value, Valid: true}
}

func ConvertOptNilDateTimeToPgTimestamptz(o api.OptNilDateTime) pgtype.Timestamptz {
	if !o.IsSet() || o.IsNull() {
		return pgtype.Timestamptz{Valid: false}
	}
	return pgtype.Timestamptz{Time: o.Value, Valid: true}
}

func ConvertOptDateTimeToPgTimestamptz(o api.OptDateTime) pgtype.Timestamptz {
	if !o.IsSet() {
		return pgtype.Timestamptz{Valid: false}
	}
	return pgtype.Timestamptz{Time: o.Value, Valid: true}
}

// We now ignore any incoming run_at, next_run and last_run values.
// They will be calculated by a worker, so we always pass null.
func NullTimestamptz() pgtype.Timestamptz {
	return pgtype.Timestamptz{Valid: false}
}

// fileExt retrieves the extension of the given file name.
func FileExt(name string) string {
	// naive
	idx := strings.LastIndex(name, ".")
	if idx >= 0 {
		return name[idx:]
	}
	return ""
}

// bytesToReader wraps raw bytes in a strings.Reader (for S3 PutObject).
func BytesToReader(b []byte) *strings.Reader {
	return strings.NewReader(string(b))
}

// optionalText converts an OptString to a pgtype.Text.
func OptionalText(opt api.OptString) pgtype.Text {
	if opt.IsSet() {
		return pgtype.Text{String: opt.Or(""), Valid: true}
	}
	return pgtype.Text{Valid: false}
}

// pgText converts a Go string to pgtype.Text
func PgText(s string) pgtype.Text {
	return pgtype.Text{String: s, Valid: true}
}

// pgInt8 converts an int to pgtype.Int8
func PgInt8(v int) pgtype.Int8 {
	return pgtype.Int8{Int64: int64(v), Valid: true}
}

// pgBool converts a bool to pgtype.Bool
func PgBool(b bool) pgtype.Bool {
	return pgtype.Bool{Bool: b, Valid: true}
}
