package models

import (
	"stock-simulator-serverless/src/selection"
	"strings"
	"time"
)

//stroage prefixes
const ChatChannelPrefix = "chat#"
const MessageIDPrefix = "message#"
const UserPrefix = "user#"
const TokenPrefix = "token#"
const TimePrefix = "time#"

type CreateUserProperties struct {
	Login          string
	Password       string
	Email          string
	InvestorType   string
	BackgroundInfo string
}

type ChangeMe struct {
	OldPassword    string
	NewPassword    string
	NewDisplayName string
	Selects        selection.User
}

type LoginUser struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type LoginUserResponse struct {
	Token Token
	User  UserStruct
}

const TimeLayout = "2006-01-02T15:04:05.999"

func TimeStringValue(t time.Time) string {
	return t.Format(TimeLayout)
}

func MustParseTimeString(s string) time.Time {
	val, err := time.Parse(TimeLayout, s)
	if err != nil {
		panic("failed to parse time: " + err.Error())
	}
	return val
}

// remove the first prefix of a key
func RemovePrefix(s string) string {
	splits := strings.Split(s, "#")
	if len(splits) == 1 {
		panic("called remove prefix on a non-prefixed word: " + s)
	}
	return splits[1]
}

func NthPos(s string, pos int) string {
	splits := strings.Split(s, "#")
	if len(splits) <= pos {
		panic("not enough parts: " + s)
	}
	return splits[pos]
}

type Trade struct {
	Company *CompanyStruct
	Count   int
	User    *UserStruct
}
