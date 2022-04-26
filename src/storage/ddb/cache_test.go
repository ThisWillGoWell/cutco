package ddb

import (
	"context"
	"cutco-camper/src/models"
	"testing"

	"github.com/stretchr/testify/assert"

)

func TestCache(t *testing.T) {
	val := Items{
		{
			"PK": {
				S: aws.String("value"),
			},
			"SK": {
				S: aws.String("skk"),
			},
		},
	}

	expr := getByPkSkStartsWith("value", "test")
	input := &dynamodb.QueryInput{
		TableName:                 aws.String("test"),
		KeyConditionExpression:    expr.KeyCondition(),
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		FilterExpression:          expr.Filter(),
	}

	ctx := context.WithValue(context.Background(), "request-id", models.NewUUID())

	hash, items := Get(ctx, input.GoString())
	assert.Nil(t, items)
	assert.NotEqual(t, "", hash)

	Add(ctx, hash, val)
	hash, items = Get(ctx, input.GoString())
	assert.NotNil(t, items)
	assert.Len(t, items, 1)

}
