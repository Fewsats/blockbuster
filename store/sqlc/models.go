// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.25.0

package sqlc

import (
	"database/sql"
	"time"
)

type MacaroonCredential struct {
	ID        int64
	TokenID   []byte
	RootKey   []byte
	CreatedAt interface{}
}

type Offer struct {
	ID             int64
	UserID         int64
	ExternalID     string
	PaymentHash    string
	PriceInCents   int64
	Currency       string
	ExpirationDate sql.NullTime
	CreatedAt      time.Time
}

type Purchase struct {
	ID             int64
	UserID         int64
	ExternalID     string
	ServiceType    string
	PriceInCents   int64
	Currency       string
	ExpirationDate sql.NullTime
	PaymentHash    string
	CreatedAt      time.Time
}

type Token struct {
	ID         int64
	Email      string
	Token      string
	Expiration time.Time
	CreatedAt  time.Time
}

type User struct {
	ID               int64
	Email            string
	LightningAddress sql.NullString
	Verified         bool
	CreatedAt        time.Time
}

type Video struct {
	ID                int64
	ExternalID        string
	UserID            int64
	Title             string
	Description       string
	CoverUrl          string
	PriceInCents      int64
	TotalViews        int64
	ThumbnailUrl      sql.NullString
	HlsUrl            sql.NullString
	DashUrl           sql.NullString
	DurationInSeconds sql.NullFloat64
	SizeInBytes       sql.NullInt64
	InputHeight       sql.NullInt64
	InputWidth        sql.NullInt64
	ReadyToStream     bool
	CreatedAt         time.Time
}
