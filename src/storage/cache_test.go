package storage

import (
	"context"
	"stock-simulator-serverless/src/models"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
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
