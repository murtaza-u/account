// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.23.0

package sqlc

import (
	"database/sql"
	"time"
)

type AuthorizationCode struct {
	ID        string
	UserID    string
	ClientID  string
	Scopes    string
	Os        sql.NullString
	Browser   sql.NullString
	ExpiresAt sql.NullTime
}

type AuthorizationHistory struct {
	UserID       string
	ClientID     string
	AuthorizedAt time.Time
}

type Client struct {
	ID                   string
	SecretHash           string
	Name                 string
	PictureUrl           sql.NullString
	AuthCallbackUrls     string
	LogoutCallbackUrls   string
	BackchannelLogoutUrl sql.NullString
	TokenExpiration      int64
	CreatedAt            time.Time
}

type Session struct {
	ID        string
	UserID    string
	ClientID  sql.NullString
	CreatedAt time.Time
	ExpiresAt time.Time
	Os        sql.NullString
	Browser   sql.NullString
}

type User struct {
	ID             string
	Email          string
	AvatarUrl      sql.NullString
	HashedPassword sql.NullString
	IsAdmin        bool
	CreatedAt      time.Time
}
