package models

import (
	"github.com/aws/aws-sdk-go/aws"
)

const PrivatePrefix = "private#"

type PrivateUserStruct struct {
	User      *UserStruct `json:"-"`
	Email     string      `json:"-"`
	Login     string      `json:"-"`
	Password  string      `json:"-"`
	VersionID string      `json:"-"`
}

func (u PrivateUserStruct) Version() string {
	return u.VersionID
}
func (u *PrivateUserStruct) VersionLoad(s string) {
	u.VersionID = s
}

// load Me by user
func (u PrivateUserStruct) PK() *string {
	return aws.String(UserPrefix + u.User.ID)
}

func (u PrivateUserStruct) SK() *string {
	return aws.String(PrivatePrefix)
}

func (u *PrivateUserStruct) Load(pk, _ string) {
	u.User = &UserStruct{ID: RemovePrefix(pk)}
}

// GSI1PK load PrivateUserStruct by login
func (u PrivateUserStruct) GSI1PK() *string {
	return aws.String(UserPrefix + u.Login)
}

func (u PrivateUserStruct) GSI1SK() *string {
	return aws.String(u.Password)
}

func (u *PrivateUserStruct) GSI1Load(pk, sk string) {
	u.Login = RemovePrefix(pk)
	u.Password = sk
}

// load PrivateUserStruct by email
func (u PrivateUserStruct) GSI2PK() *string {
	if u.Email == "" {
		return nil
	}
	return aws.String(UserPrefix + u.Email)
}

func (u PrivateUserStruct) GSI2SK() *string {
	return aws.String(PrivatePrefix)
}

func (u *PrivateUserStruct) GSI2Load(pk, _ string) {
	u.Email = RemovePrefix(pk)
}
