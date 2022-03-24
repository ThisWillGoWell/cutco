package storage

import (
	"context"
	"github.com/stretchr/testify/assert"
	"stock-simulator-serverless/src/models"
	"stock-simulator-serverless/src/selection"
	"testing"
	"time"
)

func companySeed(t *testing.T, table *DdbTable) []*models.CompanyStruct {
	ctx := context.Background()
	users := usersSeed(t, table)

	c := []*models.CompanyStruct{{
		Name:        "test-co-1",
		CreatedAt:   time.Now(),
		Symbol:      "TEST",
		Description: "desc",
		Type:        "type",
		OpenShares:  100,
		Value:       100,
		Owner:       users[0],
	}, {
		Name:        "test-co-2",
		CreatedAt:   time.Now(),
		Symbol:      "TEST2",
		Description: "desc",
		Type:        "type",
		OpenShares:  100,
		Value:       100,
		Owner:       users[1],
	}}
	for _, c := range c {
		err := table.Company.CreateCompany(ctx, c)
		assert.NoError(t, err)
		assert.NotEqual(t, "", c.ID)
	}
	return c
}

func TestCreateCompany(t *testing.T) {
	storage := NewLocalDdb()
	ctx := context.Background()
	companies := companySeed(t, storage)
	// each company should exists
	for _, c := range companies {
		res, err := storage.Company.CompanyNameAvailable(ctx, c.Name)
		assert.NoError(t, err)
		assert.False(t, res)

		res, err = storage.Company.CompanySymbolAvailable(ctx, c.Symbol)
		assert.NoError(t, err)
		assert.False(t, res)
	}
}

func TestLoadCompany(t *testing.T) {
	storage := NewLocalDdb()
	ctx := context.Background()
	companies := companySeed(t, storage)

	// load all companies
	cs, err := storage.Company.LoadCompanies(ctx, models.ReadCompaniesRequest{
		Selects: selection.Company{
			SelectInfo: true,
			Users:      &selection.User{},
		},
	})
	assert.NoError(t, err)
	assert.Len(t, cs, len(companies))

	// verify info loaded
	for _, c := range companies {
		assert.NotEmpty(t, c.ID)
		assert.NotEmpty(t, c.Description)
		assert.NotEmpty(t, c.Owner.ID)
	}
}
