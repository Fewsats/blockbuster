package auth

import (
	"time"
)

type Store interface {
	StoreToken(email, token string, expiration time.Time) error
	VerifyToken(token string) (string, error)
}
