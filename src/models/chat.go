package models

import (
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
)

type ChatChannelType string

const (
	ChatChannelTypePublic  = "public"
	ChatChannelTypePrivate = "private"
	ChatChannelTypeWhisper = "whisper"
)

type ChatChannelStruct struct {
	ID        string          `json:"-"`
	Name      string          `json:"name"`
	CreatedAt time.Time       `json:"created_at"`
	Type      ChatChannelType `json:"type"`
	// dont render lists in ddb json, not stored in the json
	Members  []*UserStruct  `json:"-"`
	Messages []*ChatMessage `json:"-"`
}

func (ChatChannelStruct) Json() {}

// address in the database
func (cc ChatChannelStruct) PK() *string {
	return aws.String(ChatChannelPrefix + cc.ID)
}
func (cc ChatChannelStruct) SK() *string {
	return aws.String(ChatChannelPrefix)
}

func (cc *ChatChannelStruct) Load(pk, _ string) {
	cc.ID = RemovePrefix(pk)
}

type ChatMessage struct {
	ID        string             `json:"-"`
	Channel   *ChatChannelStruct `json:"-"`
	Message   string             `json:"message"`
	CreatedAt time.Time          `json:"-"`
	Owner     *UserStruct        `json:"-"`
}

func (ChatMessage) Json() {}

func (cm ChatMessage) PK() *string {
	return aws.String(ChatChannelPrefix + cm.Channel.ID)
}

func (cm ChatMessage) SK() *string {
	return aws.String(fmt.Sprintf("%s%s#%s", MessageIDPrefix, TimeStringValue(cm.CreatedAt), cm.ID))
}

func (cm *ChatMessage) Load(pk, sk string) {
	cm.Channel = &ChatChannelStruct{ID: RemovePrefix(pk)}
	cm.ID = NthPos(sk, 2)
	cm.CreatedAt = MustParseTimeString(NthPos(sk, 1))
}

// load messages by user, by time and channel
func (cm ChatMessage) GSI1PK() *string {
	return aws.String(UserPrefix + cm.Owner.ID)
}

func (cm ChatMessage) GSI1SK() *string {
	return aws.String(fmt.Sprintf("%s%s#%s", MessageIDPrefix, TimeStringValue(cm.CreatedAt), cm.Channel.ID))
}

func (cm *ChatMessage) GSI1Load(pk, _ string) {
	cm.Owner = &UserStruct{ID: RemovePrefix(pk)}
}
