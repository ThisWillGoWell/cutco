package ddb

import (
	"context"
	"cutco-camper/src/models"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"golang.org/x/sync/errgroup"
)

type SingleTable struct {
	TableName string
	Client    *dynamodb.Client
}

const (
	PK = "PK"
	SK = "SK"

	JsonAttribute    = "JSON"
	VersionAttribute = "Version"
	// SK -> PK pairing with full
	GSI0Name                 = "GSI0"
	GSI1Name                 = "GSI1"
	FullProjection1IndexName = "FullProjection-1"
)

func pk(s string) string {
	return s + "-PK"
}

func sk(s string) string {
	return s + "-SK"
}

type Item interface {
	Keys() (string, string)
}

type ItemMap map[string]types.AttributeValue

type ddbItem struct {
	Item
	pk string
	sk string
}

type itemKeyPair struct {
	pk string
	sk string
}

func (i itemKeyPair) Keys() (string, string) {
	return i.pk, i.sk
}

type Version interface {
	Version() string
}

type GSI1 interface {
	GSI1() (string, string)
}

type GSI0 interface {
	GSI0() (string, string)
}

func toKey(item Item) ItemMap {
	pk, sk := item.Keys()
	return ItemMap{
		PK: &types.AttributeValueMemberS{
			Value: pk,
		},
		SK: &types.AttributeValueMemberS{
			Value: sk,
		},
	}
}

func marshal(item Item) ItemMap {
	jsonStr, err := json.Marshal(item)
	if err != nil {
		// yolo
		panic(err)
	}
	itemMap := make(ItemMap)
	itemMap[JsonAttribute] = &types.AttributeValueMemberS{
		Value: string(jsonStr),
	}

	if gsi0, ok := item.(GSI0); ok {
		pkValue, skValue := gsi0.GSI0()
		itemMap[pk(GSI0Name)], itemMap[sk(GSI0Name)] = &types.AttributeValueMemberS{
			Value: pkValue,
		}, &types.AttributeValueMemberS{
			Value: skValue,
		}

	}

	if gsi1, ok := item.(GSI1); ok {
		pkValue, skValue := gsi1.GSI1()
		itemMap[pk(GSI0Name)], itemMap[sk(GSI0Name)] = &types.AttributeValueMemberS{
			Value: pkValue,
		}, &types.AttributeValueMemberS{
			Value: skValue,
		}
	}

	if version, ok := item.(Version); ok {
		itemMap[VersionAttribute] = &types.AttributeValueMemberS{
			Value: version.Version(),
		}
	}

	return itemMap
}

type QueryIndexInput struct {
	IndexName     string
	IndexPk       string
	IndexSkEqual  string
	IndexSkPrefix string
}

func attributeKeyNames(indexName string) (string, string) {
	if indexName == "" {
		return PK, SK
	} else {
		return pk(indexName), sk(indexName)
	}
}

func toExpression(input QueryIndexInput, pkAttribute, skAttribute string) (expression.Expression, error) {

	keyCondExpr := expression.Key(pkAttribute).Equal(expression.Value(input.IndexPk))
	if input.IndexSkEqual != "" {
		keyCondExpr = keyCondExpr.And(expression.Key(skAttribute).Equal(expression.Value(input.IndexSkEqual)))
	} else if input.IndexSkPrefix != "" {
		keyCondExpr = keyCondExpr.And(expression.Key(skAttribute).BeginsWith(input.IndexSkPrefix))
	}
	return expression.NewBuilder().WithKeyCondition(keyCondExpr).Build()
}

func LoadItemByQuery(ctx context.Context, s *SingleTable, item Item, input QueryIndexInput) error {
	pkAttribute, skAttribute := attributeKeyNames(input.IndexName)
	expr, err := toExpression(input, pkAttribute, skAttribute)
	if err != nil {
		return err
	}
	var indexName *string
	if input.IndexName != "" {
		indexName = aws.String(input.IndexName)
	}
	resp, err := s.Client.Query(ctx, &dynamodb.QueryInput{
		TableName:                 aws.String(s.TableName),
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		KeyConditionExpression:    expr.KeyCondition(),
		Limit:                     aws.Int32(1),
		IndexName:                 indexName,
	})
	if err != nil {
		return err
	}
	if len(resp.Items) == 0 {
		return nil
	}
	return Unmarshal(item, resp.Items[0])
}

func LoadItemsByQuery(ctx context.Context, s *SingleTable, processItem UnmarshalFunc, input QueryIndexInput) error {
	pkAttribute, skAttribute := attributeKeyNames(input.IndexName)
	expr, err := toExpression(input, pkAttribute, skAttribute)

	var indexName *string
	if input.IndexName != "" {
		indexName = aws.String(input.IndexName)
	}
	resp, err := s.Client.Query(ctx, &dynamodb.QueryInput{
		TableName:                 aws.String(s.TableName),
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		KeyConditionExpression:    expr.KeyCondition(),
		IndexName:                 indexName,
	})
	if err != nil {
		return err
	}
	if len(resp.Items) == 0 {
		return nil
	}
	for _, item := range resp.Items {
		sk := item[skAttribute].(*types.AttributeValueMemberS)

		jsonItem := item[JsonAttribute].(*types.AttributeValueMemberS)
		if jsonItem == nil {
			return fmt.Errorf("no json attribute value")
		}
		if err := json.Unmarshal([]byte(jsonItem.Value), item); err != nil {
			return err
		}
		if err := processItem(sk.Value, []byte(jsonItem.Value)); err != nil {
			return err
		}
	}
	return nil
}

func UpdateItems(ctx context.Context, s *SingleTable, items ...Item) error {
	group, ctx := errgroup.WithContext(ctx)
	for _, item := range items {
		itemMap := marshal(item)
		pkVal, skVal := item.Keys()
		builder := expression.NewBuilder().WithKeyCondition(expression.Key(PK).Equal(expression.Value(pkVal)).And(
			expression.Key(SK).Equal(expression.Value(skVal))))

		if itemMap[VersionAttribute] != nil {
			builder = builder.WithCondition(expression.Name(VersionAttribute).Equal(expression.Value(itemMap[VersionAttribute])))
			itemMap[VersionAttribute] = &types.AttributeValueMemberS{Value: models.NewUUIDLen(10)}
		}

		var updateBuilder expression.UpdateBuilder
		first := true
		for key, value := range itemMap {
			if key == PK || key == SK {
				continue
			}
			if first {
				first = false
				updateBuilder = expression.Set(expression.Name(key), expression.Value(value))
			} else {
				updateBuilder = updateBuilder.Set(expression.Name(key), expression.Value(value))
			}
		}
		expr, err := builder.WithUpdate(updateBuilder).Build()
		if err != nil {
			return err
		}

		input := &dynamodb.UpdateItemInput{
			Key:                       toKey(item),
			TableName:                 aws.String(s.TableName),
			UpdateExpression:          expr.Update(),
			ConditionExpression:       expr.Condition(),
			ExpressionAttributeValues: expr.Values(),
			ExpressionAttributeNames:  expr.Names(),
		}
		group.Go(func() error {
			_, err := s.Client.UpdateItem(ctx, input)
			return err
		})
	}
	return group.Wait()
}

func PutItems(ctx context.Context, s *SingleTable, items ...Item) error {
	input := &dynamodb.BatchWriteItemInput{
		RequestItems: map[string][]types.WriteRequest{
			s.TableName: make([]types.WriteRequest, len(items)),
		},
	}
	for i, item := range items {
		input.RequestItems[s.TableName][i] = types.WriteRequest{
			PutRequest: &types.PutRequest{
				Item: marshal(item),
			},
		}
	}

	_, err := s.Client.BatchWriteItem(ctx, input)
	return err
}

type UnmarshalFunc func(sk string, jsonStr []byte) error

func Unmarshal(item Item, ddbMap ItemMap) error {
	jsonItem := ddbMap[JsonAttribute].(*types.AttributeValueMemberS)
	if jsonItem == nil {
		return fmt.Errorf("no json attribute value")
	}
	if err := json.Unmarshal([]byte(jsonItem.Value), item); err != nil {
		return err
	}
	return nil
}

func LoadItemByKey(ctx context.Context, s *SingleTable, item Item, pk, sk string) error {
	resp, err := s.Client.GetItem(ctx, &dynamodb.GetItemInput{
		Key:       toKey(item),
		TableName: aws.String(s.TableName),
	})

	if err != nil {
		return err
	}
	if resp.Item == nil {
		return nil
	}

	return nil
}

func unmarshalItem() {

}
