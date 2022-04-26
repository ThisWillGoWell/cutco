package models

import (
	"time"
)

const tokenPrefix = "token#"

type Token struct {
	TokenHash string    `json:"token"`
	UserID    string    `json:"user_id"`
	CreatedAt time.Time `json:"created_at"`
}

func (t Token) Keys() (string, string) {
	return tokenPrefix + t.TokenHash, t.UserID
}
