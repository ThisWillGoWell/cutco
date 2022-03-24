package logic

import (
	"context"
	"fmt"
	"stock-simulator-serverless/src/errors"
	"stock-simulator-serverless/src/models"
	"stock-simulator-serverless/src/selection"
	"stock-simulator-serverless/src/starketext"
	"stock-simulator-serverless/src/storage"
)

var (
	InvalidToken = fmt.Errorf("invalid token")
)

type Logic struct {
	User    *userLogic
	Chat    *chatLogic
	Company *companyLogic
	storage *storage.DdbTable
}

func New(table *storage.DdbTable) *Logic {
	l := &Logic{
		storage: table,
	}
	l.User = &userLogic{l}
	l.Chat = &chatLogic{l}
	l.Company = &companyLogic{l}

	return l
}

func (l *Logic) LoadAuthedUser(ctx context.Context) (*models.UserStruct, error) {
	userID, authed := starketext.AuthenticatedID(ctx)
	if !authed {
		return nil, errors.MissingAuthentication
	}

	user, err := l.storage.Users.LoadUser(ctx, models.ReadUsersRequest{
		UserID: userID,
		Selects: selection.User{
			SelectInfo: true,
		}})
	if err != nil {
		return nil, err
	}

	return user, nil
}
