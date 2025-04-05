package storage

import (
	"strings"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/shadowapi/shadowapi/backend/pkg/api"
)

// fileExt retrieves the extension of the given file name.
func fileExt(name string) string {
	// naive
	idx := strings.LastIndex(name, ".")
	if idx >= 0 {
		return name[idx:]
	}
	return ""
}

// bytesToReader wraps raw bytes in a strings.Reader (for S3 PutObject).
func bytesToReader(b []byte) *strings.Reader {
	return strings.NewReader(string(b))
}

// optionalText converts an OptString to a pgtype.Text.
func optionalText(opt api.OptString) pgtype.Text {
	if opt.IsSet() {
		return pgtype.Text{String: opt.Or(""), Valid: true}
	}
	return pgtype.Text{Valid: false}
}

// pgText converts a Go string to pgtype.Text
func pgText(s string) pgtype.Text {
	return pgtype.Text{String: s, Valid: true}
}

// pgInt8 converts an int to pgtype.Int8
func pgInt8(v int) pgtype.Int8 {
	return pgtype.Int8{Int64: int64(v), Valid: true}
}

// pgBool converts a bool to pgtype.Bool
func pgBool(b bool) pgtype.Bool {
	return pgtype.Bool{Bool: b, Valid: true}
}
