package models

import (
	"time"
)

const userPrefix = "userdao#"
const accountPrefix = "account#"

type User struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Status      string    `json:"status"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	// when was this UserStruct last active at, updated once every 5 min between requests
	LastActiveAt      time.Time `json:"last_active_at"`
	BiggestStrength   string    `json:"biggest_strength"`
	FavoriteCampSnack string    `json:"favorite_camp_snack"`
	VersionID         string    `json:"version_id"`
}

type Account struct {
	UserID            string `json:"user_id"`
	Login             string `json:"login"`
	Password          string `json:"password"`
	PasswordUpdatedOn string `json:"password_updated_on"`
	Email             string `json:"email"`
}

func (u User) Keys() (string, string) {
	return userPrefix + u.ID, InfoPrefix
}

func (u User) Version() string {
	return u.VersionID
}

func (a Account) Keys() (string, string) {
	return accountPrefix + a.UserID, InfoPrefix + accountPrefix
}

func (a Account) GSI0() (string, string) {
	return accountPrefix + a.Login, InfoPrefix
}

func (a Account) GSI1() (string, string) {
	return accountPrefix + a.Email, InfoPrefix
}
