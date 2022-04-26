package ddb

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
)

type TransactInput struct {
	PutItems []PutInput
	Updates []
}

type PutInput struct {
	Item          Item
	EnforceUnique string
}

type UpdateItem struct {
	expr
}



func (table *SingleTable) Transact(ctx context.Context, input TransactInput) error {


	transaction := dynamodb.TransactWriteItemsInput{

	}

	for _, put := range input.PutItems {
		putRequest := dynamodb.Put{
			TableName: aws.String(table.TableName),
			Item:
		}

		transaction.TransactItems = append(transaction.TransactItems, &dynamodb.TransactWriteItem{
			Put:
		})
	}

	table.Client.TransactWriteItems()
}
