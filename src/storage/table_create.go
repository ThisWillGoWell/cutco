package storage

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
)

var StarketTable = &dynamodb.CreateTableInput{
	BillingMode: aws.String(dynamodb.BillingModePayPerRequest),
	AttributeDefinitions: []*dynamodb.AttributeDefinition{
		{
			AttributeName: aws.String(PK),
			AttributeType: aws.String("S"),
		},
		{
			AttributeName: aws.String(SK),
			AttributeType: aws.String("S"),
		},
		{
			AttributeName: aws.String(GSI1PK),
			AttributeType: aws.String("S"),
		},
		{
			AttributeName: aws.String(GSI1SK),
			AttributeType: aws.String("S"),
		},
		//{
		//	AttributeName: aws.String(GSI2PK),
		//	AttributeType: aws.String("S"),
		//},
		//{
		//	AttributeName: aws.String(GSI2SK),
		//	AttributeType: aws.String("S"),
		//},
		//{
		//	AttributeName: aws.String(GSI3PK),
		//	AttributeType: aws.String("S"),
		//},
		//{
		//	AttributeName: aws.String(GSI3SK),
		//	AttributeType: aws.String("S"),
		//},
		//{
		//	AttributeName: aws.String(UpdatesPK),
		//	AttributeType: aws.String("S"),
		//},
		//{
		//	AttributeName: aws.String(UpdatesSK),
		//	AttributeType: aws.String("S"),
		//},
	},
	KeySchema: []*dynamodb.KeySchemaElement{
		{
			AttributeName: aws.String(PK),
			KeyType:       aws.String("HASH"),
		},
		{
			AttributeName: aws.String(SK),
			KeyType:       aws.String("RANGE"),
		},
	},

	GlobalSecondaryIndexes: []*dynamodb.GlobalSecondaryIndex{
		{
			IndexName: aws.String(GSI0IndexName),
			KeySchema: []*dynamodb.KeySchemaElement{
				{
					AttributeName: aws.String(SK),
					KeyType:       aws.String("HASH"),
				},
				{
					AttributeName: aws.String(PK),
					KeyType:       aws.String("RANGE"),
				},
			},
			Projection: &dynamodb.Projection{
				ProjectionType: aws.String(dynamodb.ProjectionTypeInclude),
				NonKeyAttributes: []*string{
					aws.String(VersionAttribute),
				},
			},
		},
		{
			IndexName: aws.String(GSI1IndexName),
			KeySchema: []*dynamodb.KeySchemaElement{
				{
					AttributeName: aws.String(GSI1PK),
					KeyType:       aws.String("HASH"),
				},
				{
					AttributeName: aws.String(GSI1SK),
					KeyType:       aws.String("RANGE"),
				},
			},
			Projection: &dynamodb.Projection{
				ProjectionType: aws.String(dynamodb.ProjectionTypeInclude),
				NonKeyAttributes: []*string{
					aws.String(VersionAttribute),
				},
			},
		},
		//{
		//	IndexName: aws.String(GSI2IndexName),
		//	KeySchema: []*dynamodb.KeySchemaElement{
		//		{
		//			AttributeName: aws.String(GSI2PK),
		//			KeyType:       aws.String("HASH"),
		//		},
		//		{
		//			AttributeName: aws.String(GSI2SK),
		//			KeyType:       aws.String("RANGE"),
		//		},
		//	},
		//	Projection: &dynamodb.Projection{
		//		ProjectionType: aws.String(dynamodb.ProjectionTypeKeysOnly),
		//	},
		//}, {
		//	IndexName: aws.String(GSI3IndexName),
		//	KeySchema: []*dynamodb.KeySchemaElement{
		//		{
		//			AttributeName: aws.String(GSI3PK),
		//			KeyType:       aws.String("HASH"),
		//		},
		//		{
		//			AttributeName: aws.String(GSI3SK),
		//			KeyType:       aws.String("RANGE"),
		//		},
		//	},
		//	Projection: &dynamodb.Projection{
		//		ProjectionType: aws.String(dynamodb.ProjectionTypeKeysOnly),
		//	},
		//},
		//{
		//	IndexName: aws.String(UpdatesIndexName),
		//	KeySchema: []*dynamodb.KeySchemaElement{
		//		{
		//			AttributeName: aws.String(UpdatesPK),
		//			KeyType:       aws.String("HASH"),
		//		},
		//		{
		//			AttributeName: aws.String(UpdatesSK),
		//			KeyType:       aws.String("RANGE"),
		//		},
		//	},
		//	Projection: &dynamodb.Projection{
		//		ProjectionType: aws.String(dynamodb.ProjectionTypeKeysOnly),
		//	},
		//	ProvisionedThroughput: &dynamodb.ProvisionedThroughput{
		//		ReadCapacityUnits:  aws.Int64(10),
		//		WriteCapacityUnits: aws.Int64(10),
		//	},
		//},
	},
	TableName: aws.String(TableName),
}
