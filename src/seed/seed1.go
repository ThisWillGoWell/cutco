package seed

import (
	"context"
	"stock-simulator-serverless/src/graph/model"
	"stock-simulator-serverless/src/logic"
	"stock-simulator-serverless/src/models"
	"stock-simulator-serverless/src/selection"
	"stock-simulator-serverless/src/storage"
	"time"
)

type Seed func(l *logic.Logic, s *storage.DdbTable)

var One Seed = func(l *logic.Logic, s *storage.DdbTable) {
	// create
	// make a user1
	ctx := context.Background()

	user1Token, err := l.User.Signup(ctx, UserStructTSignup(*UserList[0]))
	if err != nil {
		panic(err)
	}

	_, err = l.User.Signup(ctx, UserStructTSignup(*UserList[1]))
	if err != nil {
		panic(err)
	}
	user1Ctx := context.WithValue(ctx, "token", user1Token.Token)
	user1, err := l.User.Me(user1Ctx)
	if err != nil {
		panic(err)
	}
	// get all users
	users, err := l.User.User.UsersStruct(user1Ctx, selection.User{SelectInfo: true}, nil)
	if err != nil {
		panic(err)
	}

	// get all companies, loading the info
	companies, err := l.User.Company.GetCompaniesStruct(user1Ctx, selection.Company{SelectInfo: true}, nil)

	if err != nil {
		panic(err)
	}

	for _, c := range companies {
		_, err = l.Company.Trade(user1Ctx, model.TradeInput{
			Amount:    1,
			CompanyID: c.ID,
			Price:     c.Value,
		})
		if err != nil {
			panic(err)
		}
	}

	for _, u := range users {
		if u.ID != u.ID {
			_, err = l.Chat.SendChat(user1Ctx, model.SendChatInput{
				UserID:  &u.ID,
				Message: "test-message",
			})
		}
		break
	}

	//rite a testtoken for the user1 directly in the database
	_, err = s.Token.NewToken(ctx, models.Token{
		Token:     "testtoken",
		User:      &models.UserStruct{ID: user1.User.ID},
		CreatedAt: time.Now(),
	})
	if err != nil {
		panic(err)
	}

}
