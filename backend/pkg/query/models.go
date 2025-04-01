// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.28.0

package query

import (
	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

type Contact struct {
	UUID                    uuid.UUID          `json:"uuid"`
	UserUUID                *uuid.UUID         `json:"user_uuid"`
	InstanceUuid            *uuid.UUID         `json:"instance_uuid"`
	Status                  pgtype.Text        `json:"status"`
	Names                   []byte             `json:"names"`
	NamesSearch             pgtype.Text        `json:"names_search"`
	Last                    pgtype.Text        `json:"last"`
	First                   pgtype.Text        `json:"first"`
	Middle                  pgtype.Text        `json:"middle"`
	Birthday                pgtype.Timestamptz `json:"birthday"`
	BirthdayType            pgtype.Text        `json:"birthday_type"`
	Salary                  pgtype.Text        `json:"salary"`
	SalaryData              []byte             `json:"salary_data"`
	LastPositions           []byte             `json:"last_positions"`
	LastPositionID          pgtype.Int4        `json:"last_position_id"`
	LastPositionCompanyID   pgtype.Int4        `json:"last_position_company_id"`
	LastPositionCompanyName pgtype.Text        `json:"last_position_company_name"`
	LastPositionTitle       pgtype.Text        `json:"last_position_title"`
	LastPositionStartDate   pgtype.Timestamptz `json:"last_position_start_date"`
	LastPositionEndDate     pgtype.Timestamptz `json:"last_position_end_date"`
	LastPositionEndNow      pgtype.Bool        `json:"last_position_end_now"`
	LastPositionDescription pgtype.Text        `json:"last_position_description"`
	NoteSearch              pgtype.Text        `json:"note_search"`
	NoteKpiID               []byte             `json:"note_kpi_id"`
	Phones                  []byte             `json:"phones"`
	PhoneSearch             pgtype.Text        `json:"phone_search"`
	Phone1                  pgtype.Text        `json:"phone1"`
	Phone1Type              pgtype.Text        `json:"phone1_type"`
	Phone1Country           pgtype.Text        `json:"phone1_country"`
	Phone2                  pgtype.Text        `json:"phone2"`
	Phone2Type              pgtype.Text        `json:"phone2_type"`
	Phone2Country           pgtype.Text        `json:"phone2_country"`
	Phone3                  pgtype.Text        `json:"phone3"`
	Phone3Type              pgtype.Text        `json:"phone3_type"`
	Phone3Country           pgtype.Text        `json:"phone3_country"`
	Phone4                  pgtype.Text        `json:"phone4"`
	Phone4Type              pgtype.Text        `json:"phone4_type"`
	Phone4Country           pgtype.Text        `json:"phone4_country"`
	Phone5                  pgtype.Text        `json:"phone5"`
	Phone5Type              pgtype.Text        `json:"phone5_type"`
	Phone5Country           pgtype.Text        `json:"phone5_country"`
	Emails                  []byte             `json:"emails"`
	EmailSearch             pgtype.Text        `json:"email_search"`
	Email1                  pgtype.Text        `json:"email1"`
	Email1Type              pgtype.Text        `json:"email1_type"`
	Email2                  pgtype.Text        `json:"email2"`
	Email2Type              pgtype.Text        `json:"email2_type"`
	Email3                  pgtype.Text        `json:"email3"`
	Email3Type              pgtype.Text        `json:"email3_type"`
	Email4                  pgtype.Text        `json:"email4"`
	Email4Type              pgtype.Text        `json:"email4_type"`
	Email5                  pgtype.Text        `json:"email5"`
	Email5Type              pgtype.Text        `json:"email5_type"`
	Messengers              []byte             `json:"messengers"`
	MessengersSearch        pgtype.Text        `json:"messengers_search"`
	SkypeUuid               *uuid.UUID         `json:"skype_uuid"`
	Skype                   pgtype.Text        `json:"skype"`
	WhatsappUuid            *uuid.UUID         `json:"whatsapp_uuid"`
	Whatsapp                pgtype.Text        `json:"whatsapp"`
	TelegramUuid            *uuid.UUID         `json:"telegram_uuid"`
	Telegram                pgtype.Text        `json:"telegram"`
	WechatUuid              *uuid.UUID         `json:"wechat_uuid"`
	Wechat                  pgtype.Text        `json:"wechat"`
	LineUuid                *uuid.UUID         `json:"line_uuid"`
	Line                    pgtype.Text        `json:"line"`
	Socials                 []byte             `json:"socials"`
	SocialsSearch           pgtype.Text        `json:"socials_search"`
	LinkedinUuid            *uuid.UUID         `json:"linkedin_uuid"`
	LinkedinUrl             pgtype.Text        `json:"linkedin_url"`
	FacebookUuid            *uuid.UUID         `json:"facebook_uuid"`
	FacebookUrl             pgtype.Text        `json:"facebook_url"`
	TwitterUuid             *uuid.UUID         `json:"twitter_uuid"`
	TwitterUrl              pgtype.Text        `json:"twitter_url"`
	GithubUuid              *uuid.UUID         `json:"github_uuid"`
	GithubUrl               pgtype.Text        `json:"github_url"`
	VkUuid                  *uuid.UUID         `json:"vk_uuid"`
	VkUrl                   pgtype.Text        `json:"vk_url"`
	OdnoUuid                *uuid.UUID         `json:"odno_uuid"`
	OdnoUrl                 pgtype.Text        `json:"odno_url"`
	HhruUuid                *uuid.UUID         `json:"hhru_uuid"`
	HhruUrl                 pgtype.Text        `json:"hhru_url"`
	HabrUuid                *uuid.UUID         `json:"habr_uuid"`
	HabrUrl                 pgtype.Text        `json:"habr_url"`
	MoikrugUuid             *uuid.UUID         `json:"moikrug_uuid"`
	MoikrugUrl              pgtype.Text        `json:"moikrug_url"`
	InstagramUuid           *uuid.UUID         `json:"instagram_uuid"`
	InstagramUrl            pgtype.Text        `json:"instagram_url"`
	Social1Uuid             *uuid.UUID         `json:"social1_uuid"`
	Social1Url              pgtype.Text        `json:"social1_url"`
	Social1Type             pgtype.Text        `json:"social1_type"`
	Social2Uuid             *uuid.UUID         `json:"social2_uuid"`
	Social2Url              pgtype.Text        `json:"social2_url"`
	Social2Type             pgtype.Text        `json:"social2_type"`
	Social3Uuid             *uuid.UUID         `json:"social3_uuid"`
	Social3Url              pgtype.Text        `json:"social3_url"`
	Social3Type             pgtype.Text        `json:"social3_type"`
	Social4Uuid             *uuid.UUID         `json:"social4_uuid"`
	Social4Url              pgtype.Text        `json:"social4_url"`
	Social4Type             pgtype.Text        `json:"social4_type"`
	Social5Uuid             *uuid.UUID         `json:"social5_uuid"`
	Social5Url              pgtype.Text        `json:"social5_url"`
	Social5Type             pgtype.Text        `json:"social5_type"`
	Social6Uuid             *uuid.UUID         `json:"social6_uuid"`
	Social6Url              pgtype.Text        `json:"social6_url"`
	Social6Type             pgtype.Text        `json:"social6_type"`
	Social7Uuid             *uuid.UUID         `json:"social7_uuid"`
	Social7Url              pgtype.Text        `json:"social7_url"`
	Social7Type             pgtype.Text        `json:"social7_type"`
	Social8Uuid             *uuid.UUID         `json:"social8_uuid"`
	Social8Url              pgtype.Text        `json:"social8_url"`
	Social8Type             pgtype.Text        `json:"social8_type"`
	Social9Uuid             *uuid.UUID         `json:"social9_uuid"`
	Social9Url              pgtype.Text        `json:"social9_url"`
	Social9Type             pgtype.Text        `json:"social9_type"`
	TrackingSource          pgtype.Text        `json:"tracking_source"`
	TrackingSlug            pgtype.Text        `json:"tracking_slug"`
	CachedImg               pgtype.Text        `json:"cached_img"`
	CachedImgData           []byte             `json:"cached_img_data"`
	Crawl                   []byte             `json:"crawl"`
	DuplicateUserID         pgtype.Text        `json:"duplicate_user_id"`
	DuplicateAlternativeID  pgtype.Text        `json:"duplicate_alternative_id"`
	DuplicateReportDate     pgtype.Timestamptz `json:"duplicate_report_date"`
	EntryDate               pgtype.Timestamptz `json:"entry_date"`
	EditDate                pgtype.Timestamptz `json:"edit_date"`
	LastKpiEntryDate        pgtype.Timestamptz `json:"last_kpi_entry_date"`
}

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

type File struct {
	UUID        uuid.UUID          `json:"uuid"`
	StorageType string             `json:"storage_type"`
	StorageUuid *uuid.UUID         `json:"storage_uuid"`
	Name        string             `json:"name"`
	MimeType    pgtype.Text        `json:"mime_type"`
	Size        pgtype.Int8        `json:"size"`
	CreatedAt   pgtype.Timestamptz `json:"created_at"`
	UpdatedAt   pgtype.Timestamptz `json:"updated_at"`
}

type Message struct {
	UUID                   uuid.UUID          `json:"uuid"`
	Source                 string             `json:"source"`
	Type                   string             `json:"type"`
	ChatUuid               *uuid.UUID         `json:"chat_uuid"`
	ThreadUuid             *uuid.UUID         `json:"thread_uuid"`
	Sender                 string             `json:"sender"`
	Recipients             []string           `json:"recipients"`
	Subject                pgtype.Text        `json:"subject"`
	Body                   string             `json:"body"`
	BodyParsed             []byte             `json:"body_parsed"`
	Reactions              []byte             `json:"reactions"`
	Attachments            []byte             `json:"attachments"`
	ForwardFrom            pgtype.Text        `json:"forward_from"`
	ReplyToMessageUuid     *uuid.UUID         `json:"reply_to_message_uuid"`
	ForwardFromChatUuid    *uuid.UUID         `json:"forward_from_chat_uuid"`
	ForwardFromMessageUuid *uuid.UUID         `json:"forward_from_message_uuid"`
	ForwardMeta            []byte             `json:"forward_meta"`
	Meta                   []byte             `json:"meta"`
	CreatedAt              pgtype.Timestamptz `json:"created_at"`
	UpdatedAt              pgtype.Timestamptz `json:"updated_at"`
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
	Name      pgtype.Text        `json:"name"`
}

type Pipeline struct {
	UUID           uuid.UUID          `json:"uuid"`
	DatasourceUUID *uuid.UUID         `json:"datasource_uuid"`
	Name           string             `json:"name"`
	Type           string             `json:"type"`
	IsEnabled      bool               `json:"is_enabled"`
	Flow           []byte             `json:"flow"`
	CreatedAt      pgtype.Timestamptz `json:"created_at"`
	UpdatedAt      pgtype.Timestamptz `json:"updated_at"`
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
	UUID         uuid.UUID          `json:"uuid"`
	PipelineUuid *uuid.UUID         `json:"pipeline_uuid"`
	Type         string             `json:"type"`
	Blocklist    []string           `json:"blocklist"`
	ExcludeList  []string           `json:"exclude_list"`
	SyncAll      bool               `json:"sync_all"`
	IsEnabled    bool               `json:"is_enabled"`
	Settings     []byte             `json:"settings"`
	CreatedAt    pgtype.Timestamptz `json:"created_at"`
	UpdatedAt    pgtype.Timestamptz `json:"updated_at"`
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
	Meta      []byte             `json:"meta"`
	CreatedAt pgtype.Timestamptz `json:"created_at"`
	UpdatedAt pgtype.Timestamptz `json:"updated_at"`
}
