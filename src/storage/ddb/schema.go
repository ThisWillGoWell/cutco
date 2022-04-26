package ddb

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

func gsiGen(name string) types.GlobalSecondaryIndex {
	return types.GlobalSecondaryIndex{
		IndexName: aws.String(name),
		KeySchema: []types.KeySchemaElement{
			{
				AttributeName: aws.String(pk(name)),
				KeyType:       types.KeyTypeHash,
			}, {
				AttributeName: aws.String(sk(name)),
				KeyType:       types.KeyTypeRange,
			},
		},
		Projection: &types.Projection{
			ProjectionType:   types.ProjectionTypeInclude,
			NonKeyAttributes: []string{JsonAttribute},
		},
		ProvisionedThroughput: &types.ProvisionedThroughput{
			ReadCapacityUnits:  aws.Int64(0),
			WriteCapacityUnits: aws.Int64(0),
		},
	}
}

func TableSchema(name string) dynamodb.CreateTableInput {
	return dynamodb.CreateTableInput{
		AttributeDefinitions: []types.AttributeDefinition{
			{
				AttributeName: aws.String(PK),
				AttributeType: types.ScalarAttributeTypeS,
			}, {
				AttributeName: aws.String(SK),
				AttributeType: types.ScalarAttributeTypeS,
			}, {
				AttributeName: aws.String(pk(GSI0Name)),
				AttributeType: types.ScalarAttributeTypeS,
			}, {
				AttributeName: aws.String(sk(GSI0Name)),
				AttributeType: types.ScalarAttributeTypeS,
			}, {
				AttributeName: aws.String(pk(GSI1Name)),
				AttributeType: types.ScalarAttributeTypeS,
			}, {
				AttributeName: aws.String(sk(GSI1Name)),
				AttributeType: types.ScalarAttributeTypeS,
			},
		},
		KeySchema: []types.KeySchemaElement{
			{
				AttributeName: aws.String(PK),
				KeyType:       types.KeyTypeHash,
			}, {
				AttributeName: aws.String(SK),
				KeyType:       types.KeyTypeRange,
			},
		},
		TableName:   aws.String(name),
		BillingMode: types.BillingModePayPerRequest,
		GlobalSecondaryIndexes: []types.GlobalSecondaryIndex{
			gsiGen(GSI0Name),
			gsiGen(GSI1Name),
		},
	}

}
