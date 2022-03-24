package models

import (
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
)

type UserStruct struct {
	ID           string    `json:"-"`
	Wallet       int       `json:"-"`
	Name         string    `json:"name"`
	Description  string    `json:"description"`
	InvestorType string    `json:"investor_type"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	// when was this UserStruct last active at, updated once every 5 min between requests
	LastActiveAt time.Time      `json:"last_active_at"`
	Company      *CompanyStruct `json:"-"`
	Shares       []*ShareStruct `json:"-"`
	VersionID    string         `json:"-"`
}

func (u UserStruct) Version() string {
	return u.VersionID
}
func (u *UserStruct) VersionLoad(s string) {
	u.VersionID = s
}

// load UserStruct by ID
func (u UserStruct) PK() *string {
	return aws.String(UserPrefix + u.ID)
}

func (u UserStruct) SK() *string {
	return aws.String(UserPrefix)
}

func (u *UserStruct) Load(pk, _ string) {
	u.ID = RemovePrefix(pk)
}

func (u UserStruct) Integer() *string {
	return aws.String(fmt.Sprintf("%d", u.Wallet))
}

func (u *UserStruct) IntegerLoad(in int) {
	u.Wallet = in
}

//// load UserStruct by login
//func (u UserStruct) GSI1PK() *string {
//	return aws.String(u.Login)
//}
//
//func (u UserStruct) GSI1SK() *string {
//	return aws.String(u.Password)
//}
//
//func (u *UserStruct) GSI1Load(pk, sk string) {
//	u.Login = pk
//	u.Password = sk
//}

//// load UserStruct by email
//func (u UserStruct) GSI2PK() *string {
//	return aws.String(u.Login)
//}
//
//func (u UserStruct) GSI2SK() *string {
//	return aws.String(u.Password)
//}
//
//func (u *UserStruct) GSI2Load(pk, _ string) {
//	u.Email = pk
//}

func (u UserStruct) Json() {}

func (u UserStruct) GetUpdatedType() string {
	return UserPrefix
}
func (u UserStruct) GetUpdatedTime() time.Time {
	return u.UpdatedAt
}
