package storage

import (
	"context"
	"stock-simulator-serverless/src/models"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestChatStorage(t *testing.T) {
	ddb := NewTestingDdb(t)
	// conversation between two people
	ids := []string{
		models.NewUUID(),
		models.NewUUID(),
	}
	ctx := context.Background()
	// create a chat channel
	channel := &models.ChatChannelStruct{Name: "test"}
	err := ddb.Chat.CreateChatChannel(ctx, channel)
	assert.NoError(t, err)

	numberMessages := 10
	endTime := time.Now().Truncate(time.Second)
	startTime := endTime.Add(time.Minute * time.Duration(-1*numberMessages))

	for _, id := range ids {
		// add members to the channel
		err = ddb.Chat.AddUserToChatGroup(ctx, id, channel.ID)
		assert.NoError(t, err)
		// create one message at each time
		currentTime := startTime
		for currentTime.Before(endTime) {
			err := ddb.Chat.SaveChatMessage(ctx, &models.ChatMessage{
				ChannelID: channel.ID,
				Message:   "test-message-" + currentTime.String(),
				Owner:   &models.UserStruct{ID: id },
				CreatedAt: currentTime,
			})
			assert.NoError(t, err)
			currentTime = currentTime.Add(time.Minute)
		}

	}
	<-time.After(time.Second)
	// load chat channels for first user
	// should only be one
	chatChannels, err := ddb.Chat.loadChatChannels(ctx, models.ReadChatRequest{
		RequestingID: ids[0],
	})
	assert.NoError(t, err)
	// with each user as a memeber
	assert.Len(t, chatChannels, 1)
	assert.Equal(t, channel.ID, chatChannels[0].ID)
	channel = chatChannels[0]
	assert.Len(t, channel.Members, len(ids))

	// and should only return 5 messages
	assert.Len(t, channel.Messages, 5)
	// should return no messages after the end-time
	messages, err := ddb.Chat.LoadChatChannels(ctx, models.ReadChatRequest{
		ChannelID:    channel.ID,
		StartTime:    endTime.Add(time.Second),
		MessageLimit: 20,
	})

	assert.NoError(t, err)
	assert.Len(t, messages, 0)
	// query for all the messages in the first min of messages
	// should only return the len(ids)

	messages, err = ddb.Chat.LoadChatChannels(ctx, models.ReadChatRequest{
		ChannelID:    channel.ID,
		StartTime:    startTime.Add(time.Minute),
		MessageLimit: 30,
	})
	assert.NoError(t, err)
	assert.Len(t, messages, len(ids))

	// now getting messages from the end time should return all the messages
	messages, err = ddb.Chat.LoadChatChannels(ctx, models.ReadChatRequest{
		ChannelID:    channel.ID,
		StartTime:    endTime,
		MessageLimit: len(ids) * numberMessages,
	})
	assert.NoError(t, err)
	assert.Len(t, messages, len(ids)*numberMessages)

}
