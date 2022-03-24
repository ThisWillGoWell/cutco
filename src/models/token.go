package models

import (
	"time"

	"github.com/aws/aws-sdk-go/aws"
)

type Token struct {
	Token     string      `json:"-"`
	User      *UserStruct `json:"-"`
	CreatedAt time.Time   `json:"created_at"`
}

func (t Token) PK() *string {
	return aws.String(TokenPrefix + t.Token)
}

func (t Token) SK() *string {
	return aws.String(UserPrefix + t.User.ID)
}

func (t *Token) Load(pk, sk string) {
	t.Token = RemovePrefix(pk)
	t.User = &UserStruct{ID: RemovePrefix(sk)}
}
