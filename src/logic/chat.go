package logic

import (
	"context"
	"fmt"
	"stock-simulator-serverless/src/errors"
	"stock-simulator-serverless/src/graph/model"
	"stock-simulator-serverless/src/models"
	"stock-simulator-serverless/src/selection"
	"stock-simulator-serverless/src/starketext"
	"stock-simulator-serverless/src/storage"
	"time"
)

type chatLogic struct {
	*Logic
}

var (
	NotAllowed = fmt.Errorf("nope, cant do that")
)

// get all the messages for a userdao since a timestamp
func (chat *chatLogic) Chats(ctx context.Context, input *model.ReadChatMessagesInput) ([]*model.ChatChannel, error) {

	chats, err := chat.ChatStructs(ctx, selection.ChatChannelSelects(ctx), input)
	if err != nil {
		return nil, err
	}
	return listOfChannels(chats), err
}

func (chat *chatLogic) ChatStructs(ctx context.Context, selection selection.ChatChannel, input *model.ReadChatMessagesInput) ([]*models.ChatChannelStruct, error) {
	requset := models.ReadChatRequest{
		Selects: selection,
	}

	if input != nil {
		if input.ChannelID != nil {
			requset.ChannelID = *input.ChannelID
		}
		if input.PaginationKey != nil {
			var err error
			requset.StartTime, err = time.Parse(models.TimeLayout, *input.PaginationKey)
			if err != nil {
				return nil, errors.InvalidInputError("pagination", "failed to parse pagination key err="+err.Error())
			}
		} else if input.StartTime != nil {
			requset.StartTime = *input.StartTime
		}
		if input.MessagesLimit != nil {
			requset.MessageLimit = *input.MessagesLimit
		}
	}

	userID, ok := starketext.AuthenticatedID(ctx)
	if !ok {
		return nil, errors.MissingAuthentication
	}

	requset.RequestingID = userID
	// either read info for a single channel or for all channels the userdao is apart of
	if requset.ChannelID != "" {
		// check the userdao is part of the requested channel
		allowed, err := chat.storage.Chat.UserPartOfChatGroup(ctx, userID, requset.ChannelID)
		if err != nil {
			return nil, err
		}
		if !allowed {
			return nil, NotAllowed
		}
		// make sure the userID part of the request is empty
		requset.UserID = ""
	} else {
		requset.UserID = userID
	}

	if requset.StartTime.IsZero() {
		requset.StartTime = time.Now()
	}

	if requset.MessageLimit < 1 {
		requset.MessageLimit = 5
	} else if requset.MessageLimit > 100 {
		requset.MessageLimit = 100
	}

	channels, err := chat.storage.Chat.LoadChatChannels(ctx, requset)
	if err != nil {
		return nil, err
	}
	return channels, err
}

func (chat *chatLogic) SendChat(ctx context.Context, input model.SendChatInput) (*model.ChatMessage, error) {

	userID, ok := starketext.AuthenticatedID(ctx)
	if !ok {
		return nil, errors.MissingAuthentication
	}

	message := &models.ChatMessage{
		Message:   input.Message,
		Owner:     &models.UserStruct{ID: userID},
		CreatedAt: time.Now(),
	}

	// send to a targeted userID
	if input.UserID != nil {

		if err := chat.sendWhisperMessage(ctx, *input.UserID, message); err != nil {
			return nil, err
		}
		return messageToGql(message), nil

	}
	if input.ChannelID == nil {
		return nil, fmt.Errorf("must provide channel id or userid")
	}
	message.Channel = &models.ChatChannelStruct{ID: *input.ChannelID}
	return chat.saveChatMessage(ctx, message)
}

// load the chat channel for the whisper or make a new whisper chat channel
func (chat *chatLogic) sendWhisperMessage(ctx context.Context, toUserID string, message *models.ChatMessage) error {

	// verify the other is is legit
	_, err := chat.storage.Users.LoadUser(ctx, models.ReadUsersRequest{
		ValidateIDs: true,
		UserID:      toUserID,
	})
	if err != nil {
		if err == storage.NoEntriesFound {
			return errors.InvalidInputError("userID", "invalid userdao-id")
		}
		return errors.SomethingBadHappened("could not load to userdao", err)
	}

	message.Channel, err = chat.createOrLoadWhisperChatChannel(ctx, message.Owner.ID, toUserID)
	if err != nil {
		return err
	}

	err = chat.storage.Chat.SaveChatMessage(ctx, message)
	if err != nil {
		return err
	}
	return nil

}

// create a chat channel for a whisper conversation or return the channel
func (chat *chatLogic) createOrLoadWhisperChatChannel(ctx context.Context, fromID, toID string) (*models.ChatChannelStruct, error) {

	channelID, err := chat.storage.Chat.LoadWhisperChannelID(ctx, fromID, toID)
	// there exists a channel already, return the id
	if err == nil {
		return &models.ChatChannelStruct{
			ID: channelID,
		}, nil
	}
	channel := &models.ChatChannelStruct{
		Name:      "whisper",
		CreatedAt: time.Now(),
		Type:      models.ChatChannelTypeWhisper,
		Members:   []*models.UserStruct{{ID: fromID}, {ID: toID}},
	}
	// create a new channel
	err = chat.storage.Chat.CreateChatChannel(ctx, channel)
	if err != nil {
		return nil, err
	}
	// add each member to each chat channel and create the whisper entries
	err = chat.storage.Chat.AddUserToChatGroup(ctx, fromID, channel.ID)
	if err != nil {
		return nil, err
	}
	err = chat.storage.Chat.AddUserToChatGroup(ctx, toID, channel.ID)
	if err != nil {
		return nil, err
	}
	// create the whisper entries for quick lookup
	err = chat.storage.Chat.CreateWhisperEntries(ctx, channel.ID, fromID, toID)
	if err != nil {
		return nil, err
	}
	return channel, nil
}

func (chat *chatLogic) saveChatMessage(ctx context.Context, message *models.ChatMessage) (*model.ChatMessage, error) {
	// load chat messages
	// ensure userdao is part of group
	userID, ok := starketext.AuthenticatedID(ctx)
	if !ok {
		return nil, errors.MissingAuthentication
	}

	message.Owner = &models.UserStruct{ID: userID}
	canSend, err := chat.storage.Chat.UserPartOfChatGroup(ctx, message.Owner.ID, message.Channel.ID)
	if err != nil {
		return nil, err
	}
	if !canSend {
		return nil, NotAllowed
	}

	message.CreatedAt = time.Now()
	// save the chat message
	err = chat.storage.Chat.SaveChatMessage(ctx, message)
	if err != nil {
		return nil, err
	}
	return messageToGql(message), nil
}
