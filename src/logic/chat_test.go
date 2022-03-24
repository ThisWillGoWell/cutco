package logic

import (
	"context"
	"fmt"
	"stock-simulator-serverless/src/graph/model"
	"stock-simulator-serverless/src/models"
	"stock-simulator-serverless/src/selection"
	"stock-simulator-serverless/src/starketext"
	"stock-simulator-serverless/src/storage"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"

	"github.com/stretchr/testify/assert"
)

func TestCHatLogic(t *testing.T) {
	l := New(storage.NewTestingDdb(t))
	// signup three users
	users := []*models.UserStruct{
		{
			Login:        "user-1",
			Password:     "password",
			Description:  "hello",
			Name:         "display",
			InvestorType: "bull",
			Company: &models.CompanyStruct{
				Symbol:      "STOCK",
				Name:        "stock inc",
				Description: "",
			},
		}, {
			Login:        "user-2",
			Password:     "password",
			Description:  "hello",
			Name:         "display",
			InvestorType: "bull",
			Company: &models.CompanyStruct{
				Symbol:      "STOCK",
				Name:        "stock inc",
				Description: "",
			},
		},
	}
	// create the users
	for _, uc := range users {
		assert.NoError(t, l.User.NewUser(context.Background(), uc))
	}
	ctx := starketext.NewUserAuthed(users[0].ID)

	// create a dm between users
	channel, err := l.Chat.createOrLoadWhisperChatChannel(ctx, users[0].ID, users[1].ID)
	assert.NoError(t, err)

	// should return the same channel id
	chan2, err := l.Chat.createOrLoadWhisperChatChannel(ctx, users[0].ID, users[1].ID)
	assert.NoError(t, err)
	assert.Equal(t, channel.ID, chan2.ID)

	// send 3 messages
	for i := 0; i < 3; i++ {
		for j, user := range users {
			_, err = l.Chat.SendChat(starketext.NewUserAuthed(user.ID), model.SendChatInput{
				ChannelID: aws.String(channel.ID),
				Message:   fmt.Sprintf("test-message-%d-%d", j, i),
			})
			assert.NoError(t, err)
		}
		<-time.After(time.Millisecond * 100)
	}
	// send a final message as user 2
	_, err = l.Chat.SendChat(starketext.NewUserAuthed(users[1].ID), model.SendChatInput{
		Message:   "final-message",
		ChannelID: &channel.ID,
	})
	assert.NoError(t, err)

	// load all the messages for a user1
	channels, err := l.Chat.Chat.ChatStructs(ctx, selection.ChatChannel{
		SelectInfo:     true,
		UserSelects:    &selection.User{},
		MessageSelects: &selection.ChatMessage{},
	}, nil)

	assert.NoError(t, err)
	assert.Len(t, channels, 1)
	// should be 1 since we ignore the users
	assert.Len(t, channels[0].Members, 1)
	assert.Equal(t, channels[0].Messages[0].Message, "final-message")
}
