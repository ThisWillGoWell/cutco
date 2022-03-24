package storage

import (
	"context"
	"stock-simulator-serverless/src/models"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTokens(t *testing.T) {
	// create a new token
	userID := models.NewUUID()
	ctx := context.Background()
	ddb := NewTestingDdb(t)

	t1, err := ddb.Token.NewToken(ctx, models.Token{
		User: &models.UserStruct{ID: userID},
	})
	assert.NoError(t, err)

	// load token
	t2, err := ddb.Token.LoadToken(ctx, t1.Token)
	assert.NoError(t, err)

	assert.Equal(t, t1.User.ID, t2.User.ID)
}
