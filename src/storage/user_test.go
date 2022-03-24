package storage

import (
	"context"
	"stock-simulator-serverless/src/models"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func usersSeed(t *testing.T, table *DdbTable) []*models.UserStruct {
	now := time.Now()
	users := []*models.UserStruct{
		{
			Login:        "user1",
			Wallet:       1000,
			Email:        "me@you.com",
			Name:         "display-name",
			Password:     "password",
			Description:  "test",
			InvestorType: "bull",
			LastActiveAt: now,
			CreatedAt:    now,
			UpdatedAt:    now,
		}, {
			Login:        "user2",
			Wallet:       1000,
			Email:        "me@you.com",
			Name:         "display-name",
			Password:     "password",
			Description:  "test",
			InvestorType: "bull",
			LastActiveAt: now,
			CreatedAt:    now,
			UpdatedAt:    now,
		},
	}
	for _, u := range users {
		err := table.Users.CreateNewUser(context.Background(), u)
		assert.NoError(t, err)
	}
	return users
}

func newUser(login string) *models.UserStruct {
	if login == "" {
		login = "login"
	}
	t := time.Now()
	return &models.UserStruct{
		ID:           models.NewUUID(),
		Login:        login,
		Name:         login,
		Password:     "password",
		LastActiveAt: t,
		CreatedAt:    t,
		UpdatedAt:    t,
	}
}

func TestCreateUser(t *testing.T) {
	ddb := NewTestingDdb(t)
	ctx := context.Background()
	user := newUser("")
	err := ddb.Users.CreateNewUser(ctx, user)
	assert.NoError(t, err)

	loadedUser, err := ddb.Users.LoadUser(ctx, models.ReadUsersRequest{
		UserID: user.ID,
	})
	assert.NoError(t, err)
	assert.Equal(t, loadedUser.ID, user.ID)

}

func TestGetUserByUpdates(t *testing.T) {
	// create
	//startTime := time.Now()
	ddb := NewTestingDdb(t)
	ctx := context.Background()

	err := ddb.Users.CreateNewUser(ctx, newUser("1"))
	assert.NoError(t, err)
	err = ddb.Users.CreateNewUser(ctx, newUser("2"))
	assert.NoError(t, err)
	// pause one second and make a new user
	<-time.After(time.Second)
	//secondTime := time.Now()
	err = ddb.Users.CreateNewUser(ctx, newUser("3"))
	assert.NoError(t, err)

	// get the usersTable who have updated since we started
	allUsers, err := ddb.Users.LoadUsers(ctx, models.ReadUsersRequest{})
	assert.NoError(t, err)
	assert.Len(t, allUsers, 3)

	allUpdates, err := ddb.Users.LoadUsers(ctx, models.ReadUsersRequest{})
	assert.NoError(t, err)
	assert.Len(t, allUpdates, 3)

}

func TestDdbTable_LoadUsersByID(t *testing.T) {
	ddb := NewTestingDdb(t)
	ctx := context.Background()
	user1 := newUser("1")
	err := ddb.Users.CreateNewUser(ctx, user1)
	assert.NoError(t, err)
	user2 := newUser("2")
	err = ddb.Users.CreateNewUser(ctx, user2)
	assert.NoError(t, err)

	users, err := ddb.Users.LoadUsers(ctx, models.ReadUsersRequest{
		UserIDs:     []string{user1.ID, user2.ID, "invalid-id"},
		ValidateIDs: true,
	})
	assert.NoError(t, err)
	assert.Len(t, users, 2)
	assert.Equal(t, user1.ID, users[0].ID)
	assert.NoError(t, err)
}
