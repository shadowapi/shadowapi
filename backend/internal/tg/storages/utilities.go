package storages

import "github.com/jackc/pgx/v5/pgtype"

func ToText(value string) pgtype.Text {
	return pgtype.Text{
		Valid:  true,
		String: value,
	}
}

func ToInt8(value int64) pgtype.Int8 {
	return pgtype.Int8{
		Valid: true,
		Int64: value,
	}
}

func ToBool(value bool) pgtype.Bool {
	return pgtype.Bool{
		Valid: true,
		Bool:  value,
	}
}
