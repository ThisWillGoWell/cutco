package test

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"net/http"
	"os"
	"stock-simulator-serverless/client"
	"stock-simulator-serverless/cmd"
	"stock-simulator-serverless/src/graph/model"
	"stock-simulator-serverless/src/models"
	"testing"
	"time"
)

func TestTestSuite(t *testing.T) {
	suite.Run(t, new(TestSuite))
}

// 5 users
//

type TestUser struct {
	signup   model.SignupInput
	id       string
	client   *client.Client
	setToken func(string)
}

var ctx = context.Background()

func (u *TestUser) Signup(baseClient *client.Client) error {
	resp, err := baseClient.Signup(ctx, u.signup)
	if err != nil {
		return err
	}
	u.client = client.NewClient(&http.Client{}, baseClient.Client.BaseURL, client.AddToken(resp.Signup.Token))

	return nil
}

type TestSuite struct {
	suite.Suite
	c     *client.Client
	users []TestUser
}

// create the useres. delets the userdao if they exist
var validSingup1 = model.SignupInput{
	Login:        "user1",
	DisplayName:  "Mortis",
	Password:     "password2",
	Description:  "this is a test description",
	InvestorType: "bull",
	Email:        aws.String("me@you.com"),
	Company: &model.CreateCompanyInput{
		Name:        "chunts-blunts",
		Symbol:      "BYOB",
		Description: "this is a description.",
		Type:        "Fun",
	},
}

func (suite *TestSuite) init() []*TestUser {
	//	"https://staging.mockstarket.com/graph",
	url := os.Getenv("STARKET_URL")
	if url == "" {
		cmd.StartLocal(nil)
		url = "http://localhost:8080/graph"
	}

	suite.c = client.NewClient(
		&http.Client{},
		url,
	)
	users := []*TestUser{
		{
			signup: validSingup1,
		}, {
			signup: model.SignupInput{
				Login:        "user2",
				Email:        aws.String("hello@gmail.com"),
				DisplayName:  "pesha",
				Password:     "password2",
				Description:  "this is a test description",
				InvestorType: "bull",
				Company: &model.CreateCompanyInput{
					Name:        "Peshas-People",
					Symbol:      "BYOB",
					Description: "this is a description.",
					Type:        "Fun",
				},
			},
		},
	}
	for i := range users {
		err := users[i].Signup(suite.c)
		assert.NoError(suite.T(), err)
		resp, err := users[i].client.MyID(ctx)
		assert.NoError(suite.T(), err)
		users[i].id = resp.Me.User.ID
	}
	return users
}

func (suite *TestSuite) deleteUsers(users []*TestUser) {
	for _, u := range users {
		_, err := u.client.DeleteMe(ctx)
		assert.NoError(suite.T(), err)
		<-time.After(time.Millisecond * 10)
	}
}

func (suite *TestSuite) TestSignupUser() {
	users := suite.init()

	// make sure each userdao has an ID and company
	for _, u := range users {
		info, err := u.client.FullProfile(ctx)
		assert.NoError(suite.T(), err)
		assert.NotEqual(suite.T(), "", info.Me.User.ID)
		assert.Equal(suite.T(), u.signup.Description, info.Me.User.Description)
		assert.Equal(suite.T(), u.signup.DisplayName, info.Me.User.Name)
		assert.Equal(suite.T(), u.signup.InvestorType, info.Me.User.InvestorType)
		assert.Equal(suite.T(), u.signup.Email, info.Me.Email)

		assert.NotEqual(suite.T(), "", info.Me.User.Company.ID)
		assert.Equal(suite.T(), u.signup.Company.Name, info.Me.User.Company.Name)
		assert.Equal(suite.T(), u.signup.Company.Description, info.Me.User.Company.Description)
		assert.Equal(suite.T(), u.signup.Company.Symbol, info.Me.User.Company.Symbol)
		assert.NotEmpty(suite.T(), info.Me.User.Company.ID)
	}

	// each userdao should be able to login
	for _, u := range users {
		resp, err := suite.c.Login(ctx, model.LoginInput{
			Login:    u.signup.Login,
			Password: u.signup.Password,
		})
		assert.NoError(suite.T(), err)
		assert.NotEmpty(suite.T(), resp.Login.Token)
	}

	newUser := *users[0]

	// when I attempt to create a new company with same name it fails

	err := newUser.Signup(suite.c)
	assert.Error(suite.T(), err)
	newUser.signup.Company.Name = "newerCompany"

	// when attemtpt to create a new comapny with same symbol it fails
	err = newUser.Signup(suite.c)
	assert.Error(suite.T(), err)
	newUser.signup.Company.Symbol = "SMB"

	// when I attempt to signup with the same login it fails

	err = newUser.Signup(suite.c)
	assert.Error(suite.T(), err)
	newUser.signup.Login = "newerlogin"

	// finally it should pass
	err = newUser.Signup(suite.c)
	assert.NoError(suite.T(), err)

	users = append(users, &newUser)
	suite.deleteUsers(users)
}

func (suite *TestSuite) TestSendMessagesUser() {
	users := suite.init()

	numberMessages := 100

	// send lots of message with some time delay between them
	for _, u2 := range users[1:] {
		// send the first message to the userdao id
		mes, err := users[0].client.SendMessage(ctx, model.SendChatInput{
			UserID:  aws.String(u2.id),
			Message: "hello!",
		})
		assert.NoError(suite.T(), err)
		for i := 0; i < numberMessages/2; i++ {
			time.Sleep(time.Millisecond / 4)
			_, err = users[0].client.SendMessage(ctx, model.SendChatInput{
				ChannelID: aws.String(mes.SendChat.Channel.ID),
				Message:   fmt.Sprintf("message to %d", i),
			})
			assert.NoError(suite.T(), err)
			time.Sleep(time.Millisecond / 4)
			_, err = u2.client.SendMessage(ctx, model.SendChatInput{
				ChannelID: aws.String(mes.SendChat.Channel.ID),
				Message:   fmt.Sprintf("message back! %d", i),
			})
			assert.NoError(suite.T(), err)

		}
	}

	// now should be able to get all the chat channels that user0 is apart of
	channels, err := users[0].client.ReadChat(ctx, &model.ReadChatMessagesInput{})
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), channels.Chat, len(users)-1)

	for _, c := range channels.Chat {
		// read 100 messages off each channel, 20 times
		currentStart := models.TimeStringValue(time.Now())
		for j := 0; j < 10; j++ {
			messages, err := users[0].client.ReadChat(ctx, &model.ReadChatMessagesInput{
				ChannelID:     aws.String(c.ID),
				PaginationKey: &currentStart,
				MessagesLimit: aws.Int(numberMessages / 10),
			})
			assert.NoError(suite.T(), err)

			assert.Len(suite.T(), messages.Chat, 1, "should only return a single channel")
			assert.Len(suite.T(), messages.Chat[0].Messages, numberMessages/10)
			// update current timer
			currentStart = messages.Chat[0].Messages[len(messages.Chat[0].Messages)-1].PaginationKey
		}
		// read the final message, should be 1 left and say hello!
		messages, err := users[0].client.ReadChat(ctx, &model.ReadChatMessagesInput{
			ChannelID:     aws.String(c.ID),
			PaginationKey: &currentStart,
			MessagesLimit: aws.Int(100),
		})
		assert.NoError(suite.T(), err)
		assert.Len(suite.T(), messages.Chat, 1)
		assert.Equal(suite.T(), messages.Chat[0].Messages[0].Message, "hello!")
	}

}

func (suite *TestSuite) TestCompanies() {
	users := suite.init()
	c := users[0].client
	// load all the companies

	companies, err := c.Companines(ctx, nil)
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), companies.Companies, len(users))

}
