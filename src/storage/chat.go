package storage

import (
	"context"
	"errors"
	"fmt"
	"stock-simulator-serverless/src/models"
	"strings"
	"sync"
)

type chatTable struct {
	*DdbTable
}

func (table *chatTable) UserPartOfChatGroup(ctx context.Context, userID, channelID string) (bool, error) {
	return table.pkSkExist(ctx, UserIDPrefix+userID, ChatChannelPrefix+channelID)
}

func (table *chatTable) CreateChatChannel(ctx context.Context, channel *models.ChatChannelStruct) error {
	var err error
	for i := 0; i < 10; i++ {
		channel.ID = models.NewUUID()
		err = table.newEntry(ctx, channel)
		if err == nil {
			return nil
		}
		if err != KeyAlreadyExists {
			return err
		}
	}
	return err
}

func (table *chatTable) AddUserToChatGroup(ctx context.Context, userID, channelID string) error {
	return table.newEntry(ctx, basicEntry{pk: UserIDPrefix + userID, sk: ChatChannelPrefix + channelID})
}

func (table *chatTable) DeleteMessagesByUserID(ctx context.Context, userID string) error {
	messages, err := table.queryGSI1(ctx, queryInput{
		expr:  getByGSI1Pk(UserIDPrefix + userID),
	})
	if err != nil {
		return err
	}
	messageEntries := make([]DdbEntry, len(messages))
	for i, item := range messages {
		m := &models.ChatMessage{}
		unmarshalEntry(item,m)
		messageEntries[i] = m
	}
	err = table.DeleteAll(ctx, messageEntries )
	if err != nil {
		return err
	}
	return nil
}

func (table *chatTable) SaveChatMessage(ctx context.Context, message *models.ChatMessage) error {
	var err error
	for i := 0; i < 10; i++ {
		message.ID = models.NewUUID()
		err = table.newEntry(ctx, message)
		if err != nil {
			return nil
		}
		if err != KeyAlreadyExists {
			return err
		}
	}
	return err
}

// given two user ids, load a whisper entry
func (table *chatTable) LoadWhisperChannelID(ctx context.Context, userID1, userID2 string) (string, error) {
	items, err := table.query(ctx, queryInput{
		expr:  getByPkSkStartsWith(UserIDPrefix+userID1, UserIDPrefix+userID2+"#"+ChatChannelPrefix),
		limit: 1,
	})
	if err != nil {
		return "", err
	}
	if len(items) == 0 {
		return "", NoEntriesFound
	}
	// # live dangerously
	channelID := strings.Split(*items[0][SK].S, "#")[3]
	return channelID, nil
}

// a whisper conversation, saves under UserID1 -> UserID2# for fast lookup
// "what is the channel id of a whisper"
func (table *chatTable) CreateWhisperEntries(ctx context.Context, channelID, userID1, userID2 string) error {

	err := table.newEntry(ctx, basicEntry{pk: UserIDPrefix + userID1, sk: fmt.Sprintf("%s%s#%s%s", UserIDPrefix, userID2, ChatChannelPrefix, channelID)})
	if err != nil {
		return err
	}
	err = table.newEntry(ctx, basicEntry{pk: UserIDPrefix + userID2, sk: fmt.Sprintf("%s%s#%s%s", UserIDPrefix, userID1, ChatChannelPrefix, channelID)})
	if err != nil {
		return err
	}
	return nil
}
func (table *chatTable) LoadChatChannels(ctx context.Context, request models.ReadChatRequest) ([]*models.ChatChannelStruct, error) {
	result, err := table.loadChatChannels(ctx, request)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// load chat information
func (table *chatTable) loadChatChannels(ctx context.Context, request models.ReadChatRequest) ([]*models.ChatChannelStruct, error) {
	var err error
	var chatIDs []string

	// if a channel ID was provided, just query for that channelID
	if request.ChannelID != "" {
		// only query for information for this channelID
		chatIDs = []string{request.ChannelID}
	}

	if request.UserID != "" {
		// chat membership is stored as user#UserID / chat#ChatID
		chatIDItems, err := table.query(ctx, queryInput{expr: getByPkSkStartsWith(UserIDPrefix+request.UserID, ChatChannelPrefix)})
		if err != nil {
			return nil, err
		}
		// convert list of SK values into ids
		chatIDs = make([]string, len(chatIDItems))
		for i, item := range chatIDItems {
			chatIDs[i] = models.RemovePrefix(*item[SK].S)
		}
	}

	var chatChannels  []*models.ChatChannelStruct
	if request.Selects.SelectInfo {
		entries := make([]DdbEntry, len(chatIDs))
		for i, id := range chatIDs {
			entries[i] = &models.ChatChannelStruct{ID: id}
		}
		items, err := table.getItems(ctx, entries)
		if err != nil {
			return nil, err
		}
		chatChannels = unmarshalChatChannels(items)
	} else {
		chatChannels = make([]*models.ChatChannelStruct, len(chatIDs))
		for i, chatID := range chatIDs {
			chatChannels[i] = &models.ChatChannelStruct{ID: chatID}
		}
	}

	// load the users for the conversation if the request calls for it
	if request.Selects.UserSelects != nil {
		// load all the user IDs
		// map of user id to index of channel they are in
		usersIDSet := make(map[string][]struct {
			first  int
			second int
		}, len(chatIDs))

		userIDList := make([]string, len(chatIDs))
		// load all the members for all channels
		// save the index on where the results go
		members, err := table.loadMembersForChannels(ctx, chatIDs)
		for i, memList := range members {
			for j, m := range memList {
				if _, ok := usersIDSet[m.ID]; !ok {
					userIDList = append(userIDList, m.ID)
				}
				usersIDSet[m.ID] = append(usersIDSet[m.ID], struct {
					first  int
					second int
				}{first: i, second: j})
			}
		}

		// load users
		userList, err := table.Users.LoadUsers(ctx, models.ReadUsersRequest{
			RequestingID:       request.RequestingID,
			Selects:            *request.Selects.UserSelects,
			UserIDs:            userIDList,
			IgnoreRequestingID: true,
		})
		if err != nil {
			return nil, err
		}

		// now rematch the results
		for _, user := range userList {
			resultLocs := usersIDSet[user.ID]
			for _, result := range resultLocs {
				members[result.first][result.second] = user
			}
		}
		// and populate members
		for i := range chatChannels {
			chatChannels[i].Members = members[i]
		}
	}


	// load messages for a channel if requested
	if request.Selects.MessageSelects != nil {
		for i := range chatChannels {
			chatChannels[i].Messages, err = table.loadChatMessagesForChannel(ctx, chatChannels[i].ID, request)
		}
		if err != nil {
			return nil, err
		}
	}

	return chatChannels, nil
}

//load chat messages for a channel
func (table *DdbTable) loadChatMessagesForChannel(ctx context.Context, channelID string, request models.ReadChatRequest) ([]*models.ChatMessage, error) {
	startTimeString := models.TimeStringValue(request.StartTime)
	items, err := table.query(ctx, queryInput{
		limit: request.MessageLimit,
		expr:  getByPkSkBetween(ChatChannelPrefix+channelID, MessageIDPrefix, MessageIDPrefix+startTimeString),
		order: orderDesc,
	})
	if err != nil {
		return nil, err
	}
	messages := make([]*models.ChatMessage, len(items))
	for i, messageItem := range items {
		messages[i] = &models.ChatMessage{}
		unmarshalEntry(messageItem, messages[i])
		if request.Selects.MessageSelects.UserSelects != nil {
			messages[i].Owner, err = table.Users.LoadUser(ctx, models.ReadUsersRequest{
				RequestingID:       request.RequestingID,
				UserID:             messages[i].Owner.ID,
				Selects:            *request.Selects.MessageSelects.UserSelects,
				IgnoreRequestingID: false,
			})
			if err != nil {
				return nil, err
			}
		}
	}
	return messages, nil
}

// load members for multiple
// run loadMembersForChannel in parallel
func (table *DdbTable) loadMembersForChannels(ctx context.Context, channelID[]string)([][]*models.UserStruct, error) {
	wg := &sync.WaitGroup{}
	wg.Add(len(channelID))
	loadErrs := make([]error, len(channelID))
	hadErr := false
	members := make([][]*models.UserStruct, len(channelID))
	for i, id := range channelID {
		i := i
		id := id
		go func() {
			members[i], loadErrs[i] = table.loadMembersForChannel(ctx, id)
			if loadErrs[i] != nil {
				hadErr = true
			}
			wg.Done()
		}()
	}
	wg.Wait()

	if hadErr == true {
		errMsg := "received err during multi-load:\n"
		for _, err := range loadErrs {
			if err != nil {
				errMsg += err.Error() +"\n"
			}
		}
		return nil, errors.New(errMsg)
	}
	return members, nil
}

// load members of a chat channel
func (table *DdbTable) loadMembersForChannel(ctx context.Context, channelID string) ([]*models.UserStruct, error) {
	// use GSI0 SK->PK to find members,
	userItems, err := table.queryGSI0(ctx, queryInput{expr: getBySkPkStartsWith(ChatChannelPrefix+channelID, UserIDPrefix)})
	if err != nil {
		return nil, err
	}
	return unmarshalUsers(userItems), nil
}


func unmarshalChatChannel(item Item ) *models.ChatChannelStruct{
	c := &models.ChatChannelStruct{}
	unmarshalEntry(item, c)
	return c
}

func unmarshalChatChannels(items Items) []*models.ChatChannelStruct {
	chatChannels := make([]*models.ChatChannelStruct, len(items))
	for i, item := range items {
		chatChannels[i] = unmarshalChatChannel(item)
	}
	return chatChannels
}