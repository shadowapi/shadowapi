package handler

type Empty = struct{}

type SignInRequest struct {
	Body SignInRequestBody
}

type SignInRequestBody struct {
	Phone string `json:"phone" example:"+16505551234" doc:"Phone number in international format" validate:"required,e164"`
}

type SelfRequest struct {
	ID int64 `path:"id" doc:"Session id"`
}

type SelfResponse struct {
	Body User
}

type SignInResponse struct {
	Body SignInResponseBody
}

type SignInResponseBody struct {
	SessionID     int64  `json:"session_id" doc:"Session id"`
	PhoneCodeHash string `json:"phone_code_hash" doc:"SMS code ID"`
	Timeout       int64  `json:"timeout" doc:"Timeout for reception of the phone code"`
	NextType      string `json:"next_type" doc:"Phone code type that will be sent next"`
}

type SendCodeRequest struct {
	ID   int64 `path:"id" doc:"Session id"`
	Body SendCodeRequestBody
}

type SendCodeRequestBody struct {
	PhoneCodeHash string `json:"phone_code_hash" doc:"SMS code ID"`
	Password      string `json:"password" doc:"Required if 2FA is on"`
	Code          string `json:"code" doc:"SMS code"`
}

type SendCodeResponse struct {
	Body User
}

type User struct {
	ID        int64  `json:"id"         doc:"User ID in telegram"`
	Username  string `json:"username"   doc:"Username in telegram"`
	FirstName string `json:"first_name" doc:"First name"`
	LastName  string `json:"last_name"  doc:"Last name"`
	Phone     string `json:"phone"      doc:"Users phone"`
}

type SessionListResponse struct {
	Body SessionListResponseBody
}

type SessionListResponseBody struct {
	Total    int64     `json:"total"    doc:"Total number of sessions for current account"`
	Sessions []Session `json:"sessions" doc:"List of available sessions"`
}

type Session struct {
	ID          int64  `json:"id"                    doc:"Session id"`
	Phone       string `json:"phone"                 doc:"Session phone number"`
	Description string `json:"description,omitempty" doc:"Optional description"`
	UpdatedAt   string `json:"updated_at"            doc:"Last time when session was updated"`
	CreatedAt   string `json:"created_at"            doc:"Time when session was created"`
}
