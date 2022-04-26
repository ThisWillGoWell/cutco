package seed

import (
	"context"
	"fmt"
	"math/rand"
	"stock-simulator-serverless/src/graph/model"
	"stock-simulator-serverless/src/logic"
	"stock-simulator-serverless/src/models"
	"stock-simulator-serverless/src/selection"
	"stock-simulator-serverless/src/starketext"
	"stock-simulator-serverless/src/storage"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
)

var Two Seed = func(l *logic.Logic, s *storage.DdbTable) {
	// make all the users
	// make all the companies not made
	ctx := context.Background()

	for i, u := range UserList {
		err := l.User.NewUser(ctx, u)
		if err != nil {
			panic(err)
		}

		UserList[i].User.Wallet = 100000000
		UserList[i].User.VersionID, err = s.UpdateEntry(ctx, UserList[i].User)
		if err != nil {
			panic(err)
		}
	}

	for i, c := range CompanyList {
		if i >= len(UserList) {
			company, err := l.Company.CreateCompany(ctx, *c, true)
			if err != nil {
				panic(err)
			}
			CompanyList[i] = company
		} else {
			var err error
			CompanyList[i], err = s.Company.LoadCompany(ctx, models.ReadCompaniesRequest{
				OwnerID: UserList[i].User.ID,
				Selects: selection.Company{
					SelectInfo: true,
				},
			})
			if err != nil {
				panic(err)
			}
		}
	}

	// add 5 shares of each company to each userdao
	//rite a testtoken for the user1 directly in the database
	_, err := s.Token.NewToken(ctx, models.Token{
		Token:     "testtoken",
		User:      &models.UserStruct{ID: UserList[0].User.ID},
		CreatedAt: time.Now(),
	})
	if err != nil {
		panic(err)
	}

	// buy the shares and send some whisper messages
	for i := range UserList {
		ctx = starketext.NewUserAuthed(UserList[i].User.ID)
		for _, u := range UserList {
			if u.User.ID != UserList[i].User.ID {
				// do I have a whisper already?
				_, err := l.Chat.SendChat(ctx, model.SendChatInput{
					UserID:  aws.String(u.User.ID),
					Message: fmt.Sprintf("hello %s! my name is %s would you like to invest in my company?", u.User.Name, UserList[i].User.Name)})
				if err != nil {
					panic(err)
				}
			}
		}

		for j := 0; j < 20; j++ {

			c := CompanyList[rand.Intn(len(CompanyList))]

			_, err = l.Company.Trade(ctx, model.TradeInput{
				CompanyID: c.ID,
				Amount:    rand.Intn(5) + 1,
				Price:     c.Value,
			})
			if err != nil {
				panic(err)
			}
		}
	}
}
