package storage

import (
	"context"
	"errors"
	"fmt"
	errors2 "stock-simulator-serverless/src/errors"
	"stock-simulator-serverless/src/models"
	"time"

	"github.com/aws/aws-sdk-go/service/dynamodb/expression"
)

type tokenTable struct {
	*DdbTable
}

func (table *tokenTable) NewToken(ctx context.Context, token models.Token) (models.Token, error) {
	token.CreatedAt = time.Now()
	var err error
	// use a predefined value if there at first
	if token.Token == "" {
		token.Token = models.NewUUIDLen(32)
	}

	for i := 0; i < 10; i++ {
		err = table.newEntry(ctx, token)
		if err == nil {
			return token, nil
		}
		// only continue on KeyAlreadyExists
		if err != KeyAlreadyExists {
			return models.Token{}, err
		}
		token.Token = models.NewUUIDLen(32)
	}

	return models.Token{}, fmt.Errorf("failed to find a key to insert into")
}

func (table *tokenTable) LoadToken(ctx context.Context, token string) (models.Token, error) {
	built, err := expression.NewBuilder().WithKeyCondition(
		expression.KeyEqual(expression.Key(PK), expression.Value(models.TokenPrefix+token))).Build()
	if err != nil {
		return models.Token{}, err
	}

	items, err := table.query(ctx, queryInput{expr: built})
	if err != nil {
		return models.Token{}, err
	}
	if len(items) == 0 {
		return models.Token{}, errors2.NoEntriesFound
	}

	if len(items) != 1 {
		return models.Token{}, errors.New("returned the wrong number of entries")
	}

	t := models.Token{}
	unmarshalEntry(items[0], &t)

	return t, nil
}
