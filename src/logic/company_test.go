package logic

import (
	"context"
	"stock-simulator-serverless/src/graph/model"
	"stock-simulator-serverless/src/models"
	"stock-simulator-serverless/src/selection"
	"stock-simulator-serverless/src/starketext"
	"stock-simulator-serverless/src/storage"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCompanyLogic_Trade(t *testing.T) {
	logic := New(storage.NewLocalDdb())
	user := &models.UserStruct{Login: "login", Password: "pass", Wallet: 100}

	err := logic.storage.Users.CreateNewUser(context.Background(), user)
	assert.NoError(t, err)

	ctx := starketext.NewUserAuthed(user.ID)

	company := &models.CompanyStruct{
		Name:        "name",
		Description: "desc",
		Value:       1000,
		OpenShares:  11,
	}

	err = logic.storage.Company.CreateCompany(ctx, company)
	assert.NoError(t, err)

	// should fail because too expensive
	_, err = logic.Company.Trade(ctx, model.TradeInput{
		CompanyID: company.ID,
		Amount:    10,
		Price:     1000,
	})
	assert.Error(t, err)
	company.Value = 10
	// make it cheaper
	_, err = logic.storage.UpdateEntry(ctx, company)
	assert.NoError(t, err)
	// should pass this time
	_, err = logic.Company.Trade(ctx, model.TradeInput{
		CompanyID: company.ID,
		Amount:    10,
		Price:     company.Value,
	})
	assert.NoError(t, err)

	// a share should exist
	share, err := logic.storage.Company.LoadShare(ctx, models.ReadSharesRequest{
		CompanyID: company.ID,
		HolderID:  user.ID,
		Selects: selection.Share{
			SelectInfo: true,
		},
	})
	assert.NoError(t, err)
	assert.Equal(t, 10, share.Count)
	// the user should be broke
	user, err = logic.storage.Users.LoadUser(ctx, models.ReadUsersRequest{
		UserID: user.ID,
		Selects: selection.User{
			SelectInfo: true,
		},
	})
	assert.NoError(t, err)
	assert.Equal(t, user.Wallet, 0)

}
