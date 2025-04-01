package handler

import (
	"fmt"
	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/shadowapi/shadowapi/backend/pkg/api"
)

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

func uToBytes(u uuid.UUID) [16]byte {
	var arr [16]byte
	copy(arr[:], u.Bytes())
	return arr
}
