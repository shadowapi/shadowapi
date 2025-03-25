// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.26.0

package query

import (
	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

type Datasource struct {
	UUID      uuid.UUID          `json:"uuid"`
	UserUUID  *uuid.UUID         `json:"user_uuid"`
	Name      string             `json:"name"`
	Type      string             `json:"type"`
	IsEnabled bool               `json:"is_enabled"`
	Provider  string             `json:"provider"`
	Settings  []byte             `json:"settings"`
	CreatedAt pgtype.Timestamptz `json:"created_at"`
	UpdatedAt pgtype.Timestamptz `json:"updated_at"`
}

type Oauth2Client struct {
	ID        string             `json:"id"`
	Name      string             `json:"name"`
	Provider  string             `json:"provider"`
	Secret    string             `json:"secret"`
	CreatedAt pgtype.Timestamptz `json:"created_at"`
	UpdatedAt pgtype.Timestamptz `json:"updated_at"`
}

type Oauth2State struct {
	UUID       uuid.UUID          `json:"uuid"`
	ClientName string             `json:"client_name"`
	ClientID   string             `json:"client_id"`
	State      []byte             `json:"state"`
	CreatedAt  pgtype.Timestamptz `json:"created_at"`
	UpdatedAt  pgtype.Timestamptz `json:"updated_at"`
	ExpiredAt  pgtype.Timestamptz `json:"expired_at"`
}

type Oauth2Subject struct {
	UUID       uuid.UUID          `json:"uuid"`
	UserUUID   *uuid.UUID         `json:"user_uuid"`
	ClientName string             `json:"client_name"`
	ClientID   string             `json:"client_id"`
	Token      []byte             `json:"token"`
	CreatedAt  pgtype.Timestamptz `json:"created_at"`
	UpdatedAt  pgtype.Timestamptz `json:"updated_at"`
	ExpiredAt  pgtype.Timestamptz `json:"expired_at"`
}

type Oauth2Token struct {
	UUID      uuid.UUID          `json:"uuid"`
	ClientID  string             `json:"client_id"`
	Token     []byte             `json:"token"`
	CreatedAt pgtype.Timestamptz `json:"created_at"`
	UpdatedAt pgtype.Timestamptz `json:"updated_at"`
}

type Pipeline struct {
	UUID      uuid.UUID          `json:"uuid"`
	Name      string             `json:"name"`
	Flow      []byte             `json:"flow"`
	CreatedAt pgtype.Timestamptz `json:"created_at"`
	UpdatedAt pgtype.Timestamptz `json:"updated_at"`
}

type PipelineEntry struct {
	UUID         uuid.UUID          `json:"uuid"`
	PipelineUuid *uuid.UUID         `json:"pipeline_uuid"`
	ParentUuid   *uuid.UUID         `json:"parent_uuid"`
	Type         string             `json:"type"`
	Params       []byte             `json:"params"`
	CreatedAt    pgtype.Timestamptz `json:"created_at"`
	UpdatedAt    pgtype.Timestamptz `json:"updated_at"`
}

type Storage struct {
	UUID      uuid.UUID          `json:"uuid"`
	Name      string             `json:"name"`
	Type      string             `json:"type"`
	IsEnabled bool               `json:"is_enabled"`
	Settings  []byte             `json:"settings"`
	CreatedAt pgtype.Timestamptz `json:"created_at"`
	UpdatedAt pgtype.Timestamptz `json:"updated_at"`
}

type SyncPolicy struct {
	UUID        uuid.UUID          `json:"uuid"`
	UserID      pgtype.UUID        `json:"user_id"`
	Service     string             `json:"service"`
	Blocklist   []string           `json:"blocklist"`
	ExcludeList []string           `json:"exclude_list"`
	SyncAll     bool               `json:"sync_all"`
	Settings    []byte             `json:"settings"`
	CreatedAt   pgtype.Timestamptz `json:"created_at"`
	UpdatedAt   pgtype.Timestamptz `json:"updated_at"`
}

type TgAccount struct {
	ID       int64       `json:"id"`
	Username pgtype.Text `json:"username"`
}

type TgCachedChannel struct {
	ID          int64       `json:"id"`
	Title       pgtype.Text `json:"title"`
	Username    pgtype.Text `json:"username"`
	Broadcast   pgtype.Bool `json:"broadcast"`
	Forum       pgtype.Bool `json:"forum"`
	Megagroup   pgtype.Bool `json:"megagroup"`
	Raw         []byte      `json:"raw"`
	RawFull     []byte      `json:"raw_full"`
	FkSessionID int64       `json:"fk_session_id"`
}

type TgCachedChat struct {
	ID          int64       `json:"id"`
	Title       pgtype.Text `json:"title"`
	Raw         []byte      `json:"raw"`
	RawFull     []byte      `json:"raw_full"`
	FkSessionID int64       `json:"fk_session_id"`
}

type TgCachedUser struct {
	ID          int64       `json:"id"`
	FirstName   pgtype.Text `json:"first_name"`
	LastName    pgtype.Text `json:"last_name"`
	Username    pgtype.Text `json:"username"`
	Phone       pgtype.Text `json:"phone"`
	Raw         []byte      `json:"raw"`
	RawFull     []byte      `json:"raw_full"`
	FkSessionID int64       `json:"fk_session_id"`
}

type TgPeer struct {
	ID          int64       `json:"id"`
	FkSessionID int64       `json:"fk_session_id"`
	PeerType    string      `json:"peer_type"`
	AccessHash  pgtype.Int8 `json:"access_hash"`
}

type TgPeersChannel struct {
	ID          int64 `json:"id"`
	FkSessionID int64 `json:"fk_session_id"`
	Pts         int64 `json:"pts"`
}

type TgPeersUser struct {
	ID          int64       `json:"id"`
	FkSessionID int64       `json:"fk_session_id"`
	Phone       pgtype.Text `json:"phone"`
}

type TgSession struct {
	ID           int64              `json:"id"`
	Phone        string             `json:"phone"`
	AccountID    int64              `json:"account_id"`
	Session      []byte             `json:"session"`
	ContactsHash pgtype.Int8        `json:"contacts_hash"`
	Description  pgtype.Text        `json:"description"`
	CreatedAt    pgtype.Timestamptz `json:"created_at"`
	UpdatedAt    pgtype.Timestamptz `json:"updated_at"`
}

type TgSessionsState struct {
	ID   int64 `json:"id"`
	Pts  int64 `json:"pts"`
	Qts  int64 `json:"qts"`
	Date int64 `json:"date"`
	Seq  int64 `json:"seq"`
}

type User struct {
	UUID      uuid.UUID          `json:"uuid"`
	Email     string             `json:"email"`
	Password  string             `json:"password"`
	FirstName string             `json:"first_name"`
	LastName  string             `json:"last_name"`
	IsEnabled bool               `json:"is_enabled"`
	IsAdmin   bool               `json:"is_admin"`
	CreatedAt pgtype.Timestamptz `json:"created_at"`
	UpdatedAt pgtype.Timestamptz `json:"updated_at"`
}
