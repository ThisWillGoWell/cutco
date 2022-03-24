package errors

import (
	"errors"
	"fmt"
)

var (
	LoginNameTaken   = errors.New("login name already taken")
	DisplayNameTaken = errors.New("display name taken")
	WrongPassword    = errors.New("incorrect password")
)

func UnknownError(err error) error {
	return fmt.Errorf("somthing unknown happened! help! err=%v", err)
}

type Error struct {
	message string
}

func (e *Error) Error() string {
	return e.message
}

func MissingInputError(name string) *Error {
	return &Error{
		message: fmt.Sprintf("operation requires input: %s", name),
	}
}

func InvalidInputError(name, message string) *Error {
	return &Error{
		message: fmt.Sprintf("invalid input: %s, %s", name, message),
	}
}

func SomethingBadHappened(msg string, err error) *Error {
	return &Error{
		message: fmt.Sprintf("ohhhh noooo, something bad happened!!! msg=[%s] err=[%v]", msg, err),
	}
}

var NoEntriesFound = &Error{"No entries found"}
var MissingAuthentication = &Error{"missing authentication, check the Authorization header "}

func Equal(err error, Err *Error) bool {
	if err == nil {
		return false
	}
	return err.Error() == Err.Error()
}
