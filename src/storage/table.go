package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"stock-simulator-serverless/src/models"
	"stock-simulator-serverless/src/starketext"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go/aws/awserr"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/expression"
)

var (
	TableName = "starket-table"

	PK            = "PK"
	SK            = "SK"
	GSI0IndexName = "GSI0"

	VersionAttribute  = "Version"
	IntegerAttribute  = "Integer"
	Integer2Attribute = "Integer2"
	GSI1IndexName     = "GSI1"
	GSI1PK            = "GSI1PK"
	GSI1SK            = "GSI1SK"

	GSI2IndexName = "GSI2"
	GSI2PK        = "GSI2PK"
	GSI2SK        = "GSI2SK"

	GSI3IndexName = "GSI3"
	GSI3PK        = "GSI3PK"
	GSI3SK        = "GSI3SK"

	UpdatesIndexName = "updates"
	UpdatesPK        = "type"
	UpdatesSK        = "time"

	LastUpdatedAtIndex = "LastUpdatedAt"
	JsonValueAttribute = "JsonValue"
)

const (
	ChatChannelPrefix = models.ChatChannelPrefix
	MessageIDPrefix   = models.MessageIDPrefix
	UserIDPrefix      = models.UserPrefix
	TokenPrefix       = models.TokenPrefix
)

type Item map[string]*dynamodb.AttributeValue
type Items []map[string]*dynamodb.AttributeValue

// return the keys of the item
// all items should have sk and pk attached
func (i Item) Keys() (string, string) {
	if _, ok := i[PK]; !ok {
		panic("PK missing on item")
	}
	pk := i[PK].S
	if pk == nil {
		panic("PK nil on item")
	}

	if _, ok := i[SK]; !ok {
		panic("SK missing on item")
	}
	sk := i[SK].S
	if sk == nil {
		panic("SK nil on item")
	}
	return *pk, *sk
}

func (i Item) CacheKey() string {
	pk, sk := i.Keys()
	return fmt.Sprintf("%s-%s", pk, sk)
}
func (i Item) Version() bool {
	if val, ok := i[VersionAttribute]; !ok || val.S == nil {
		return false
	}
	return true
}
func (i Item) VersionValue() string {
	return *i[VersionAttribute].S
}

func (i Item) Integer() bool {
	if val, ok := i[IntegerAttribute]; !ok || val.N == nil {
		return false
	}
	return true
}
func (i Item) IntegerValue() int {
	intVal, _ := strconv.ParseInt(*i[IntegerAttribute].N, 10, 64)
	return int(intVal)
}

func (i Item) Integer2() bool {
	if val, ok := i[Integer2Attribute]; !ok || val.N == nil {
		return false
	}
	return true
}
func (i Item) Integer2Value() int {
	intVal, _ := strconv.ParseInt(*i[Integer2Attribute].N, 10, 64)
	return int(intVal)
}

func (i Item) GSI1() bool {
	if val, ok := i[GSI1PK]; !ok || val.S == nil {
		return false
	}
	if val, ok := i[GSI1SK]; !ok || val.S == nil {
		return false
	}
	return true
}

func (i Item) GSI1Values() (string, string) {
	return *i[GSI1PK].S, *i[GSI1SK].S
}

func (i Item) GSI2() bool {
	if val, ok := i[GSI2PK]; !ok || val.S == nil {
		return false
	}
	if val, ok := i[GSI2SK]; !ok || val.S == nil {
		return false
	}
	return true
}

func (i Item) GSI2Values() (string, string) {
	return *i[GSI2PK].S, *i[GSI2SK].S
}

func (i Item) GSI3() bool {
	if val, ok := i[GSI3PK]; !ok || val.S == nil {
		return false
	}
	if val, ok := i[GSI3SK]; !ok || val.S == nil {
		return false
	}
	return true
}

func (i Item) GSI3Values() (string, string) {
	return *i[GSI3PK].S, *i[GSI3SK].S
}

func (i Item) Json() bool {
	if val, ok := i[JsonValueAttribute]; !ok || val.S == nil {
		return false
	}
	return true
}

func (i Item) JsonValue() string {
	return *i[JsonValueAttribute].S
}

type DdbTable struct {
	ddb     *dynamodb.DynamoDB
	Name    string
	Users   usersTable
	Company companyTable
	Chat    chatTable
	Token   tokenTable
}

func New(tableName string, ddb *dynamodb.DynamoDB) *DdbTable {
	t := &DdbTable{ddb: ddb, Name: tableName}
	t.Users = usersTable{t}
	t.Company = companyTable{t}
	t.Chat = chatTable{t}
	t.Token = tokenTable{t}
	return t
}

type DdbEntry interface {
	PK() *string
	SK() *string
}

type Load interface {
	Load(pk, sk string)
}

type basicEntry struct {
	pk string
	sk string
}

func (b basicEntry) PK() *string {
	return aws.String(b.pk)
}
func (b basicEntry) SK() *string {
	return aws.String(b.sk)
}

func (b *basicEntry) Load(pk, sk string) {
	b.sk = sk
	b.pk = pk
}

type basicTimeEntry struct {
	basicEntry
	timePk string
	time   time.Time
}

func (b basicTimeEntry) TimePk() string {
	return b.timePk
}

func (b basicTimeEntry) Time() string {
	return models.TimeStringValue(b.time)
}

func (b *basicTimeEntry) TimeIndexLoad(pk, t string) {
	b.pk = pk
	b.time = models.MustParseTimeString(t)
}

type LastUpdatedIndex interface {
	GetUpdatedType() string
	GetUpdatedTime() time.Time
}

type GSI1 interface {
	GSI1PK() *string
	GSI1SK() *string
}

type GSI1Load interface {
	GSI1Load(pk, sk string)
}

type Version interface {
	Version() string
}

type VersionLoad interface {
	VersionLoad(string)
}

type Integer interface {
	Integer() *string
}

type IntegerLoad interface {
	IntegerLoad(int)
}

type Integer2 interface {
	Integer2() *string
}

type Integer2Load interface {
	Integer2Load(int)
}

type GSI2 interface {
	GSI2PK() *string
	GSI2SK() *string
}
type GSI2Load interface {
	GSI2Load(pk, sk string)
}

type GSI3 interface {
	GSI3PK() *string
	GSI3SK() *string
}

type GSI3Load interface {
	GSI3Load(pk, sk string)
}

type Json interface {
	// should the item be json
	Json()
}

// unmarshal a ddb item into a pointer provided
func unmarshalEntry(item Item, target Load) {

	//load the PK and SK
	target.Load(item.Keys())

	// if the target has GSI1 implemented and the item contains the values
	if gsi, ok := target.(GSI1Load); ok && item.GSI1() {
		gsi.GSI1Load(item.GSI1Values())
	}

	// if the target has GSI2 implemented and the item contains the values
	if gsi, ok := target.(GSI2Load); ok && item.GSI2() {
		gsi.GSI2Load(item.GSI2Values())
	}

	// if the target has GSI3 implemented and the item contains the values
	if gsi, ok := target.(GSI3Load); ok && item.GSI3() {
		gsi.GSI3Load(item.GSI3Values())
	}

	// if the target has any values stored in the json blob, marshal them
	if item.Json() {
		err := json.Unmarshal([]byte(item.JsonValue()), target)
		if err != nil {
			panic(err)
		}
	}

	if version, ok := target.(VersionLoad); ok && item.Version() {
		version.VersionLoad(item.VersionValue())
	}

	if integer, ok := target.(IntegerLoad); ok && item.Integer() {
		integer.IntegerLoad(item.IntegerValue())
	}

	if integer, ok := target.(Integer2Load); ok && item.Integer2() {
		integer.Integer2Load(item.Integer2Value())
	}

}

func marshalEntry(entry DdbEntry) Item {

	m := map[string]*dynamodb.AttributeValue{
		PK: {S: entry.PK()},
		SK: {S: entry.SK()},
	}

	if _, ok := entry.(Json); ok {
		jsonValue, err := json.Marshal(entry)
		// should never happen, but just in case
		if err != nil {
			panic("failed to marshal json entry: " + err.Error())
		}
		m[JsonValueAttribute] = &dynamodb.AttributeValue{S: aws.String(string(jsonValue))}
	}

	// attach GSI1 indexes if implements interface and both pk/sk are nil
	if gsi, ok := entry.(GSI1); ok {
		gsiPK, gsiSK := gsi.GSI1PK(), gsi.GSI1SK()
		if gsiPK != nil && gsiSK != nil {
			m[GSI1PK] = &dynamodb.AttributeValue{S: gsiPK}
			m[GSI1SK] = &dynamodb.AttributeValue{S: gsiSK}
		}
	}

	// attach GSI1 indexes if implements interface and both pk/sk are nil
	if gsi, ok := entry.(GSI2); ok {
		gsiPK, gsiSK := gsi.GSI2PK(), gsi.GSI2SK()
		if gsiPK != nil && gsiSK != nil {
			m[GSI2PK] = &dynamodb.AttributeValue{S: gsiPK}
			m[GSI2SK] = &dynamodb.AttributeValue{S: gsiSK}
		}
	}

	// attach GSI1 indexes if implements interface and both pk/sk are nil
	if gsi, ok := entry.(GSI3); ok {
		gsiPK, gsiSK := gsi.GSI3PK(), gsi.GSI3SK()
		if gsiPK != nil && gsiSK != nil {
			m[GSI3PK] = &dynamodb.AttributeValue{S: gsiPK}
			m[GSI3SK] = &dynamodb.AttributeValue{S: gsiSK}
		}
	}

	if v, ok := entry.(Version); ok {
		if v.Version() == "" {
			m[VersionAttribute] = &dynamodb.AttributeValue{S: aws.String("x")}
		} else {
			m[VersionAttribute] = &dynamodb.AttributeValue{S: aws.String(models.NewUUID())}
		}
	}

	if integer, ok := entry.(Integer); ok {
		m[IntegerAttribute] = &dynamodb.AttributeValue{N: integer.Integer()}
	}
	if integer, ok := entry.(Integer2); ok {
		m[Integer2Attribute] = &dynamodb.AttributeValue{N: integer.Integer2()}
	}

	return m
}

func (table *DdbTable) getItems(ctx context.Context, entries []DdbEntry) (Items, error) {
	if len(entries) > 100 {
		return nil, fmt.Errorf("too many requests ")
	}
	keys := &dynamodb.KeysAndAttributes{}
	keys.Keys = make([]map[string]*dynamodb.AttributeValue, len(entries))

	for i, e := range entries {
		keys.Keys[i] = keyFromEntry(e)
	}

	input := &dynamodb.BatchGetItemInput{
		RequestItems: map[string]*dynamodb.KeysAndAttributes{
			table.Name: keys,
		},
	}
	inputHash, items := Get(ctx, input.GoString())

	if items != nil {
		return items, nil
	}

	result, err := table.ddb.BatchGetItemWithContext(ctx, input)
	if err != nil {
		return nil, nil
	}
	res := result.Responses[table.Name]
	log := starketext.Logger(ctx)
	for _, item := range res {
		log.Infow("get-items", "pk", *item[PK].S, "sk", *item[SK].S)
	}

	Add(ctx, inputHash, res)
	return res, nil
}

func (table *DdbTable) getItem(ctx context.Context, entry DdbEntry) (Item, error) {
	input := &dynamodb.GetItemInput{
		TableName: aws.String(table.Name),
		Key:       keyFromEntry(entry),
	}

	inputHash, items := Get(ctx, input.GoString())
	if items != nil {
		return items[0], nil
	}
	item, err := table.ddb.GetItemWithContext(ctx, input)
	if err != nil {
		return nil, err
	}
	starketext.Logger(ctx).Infow("get-item", "pk", *item.Item[PK].S, "sk", *item.Item[SK].S)

	AddSingle(ctx, inputHash, item.Item)
	return item.Item, nil
}

type orderType int

var (
	noOrderDefined orderType = 0
	orderDesc      orderType = 1
	orderAsc       orderType = 2
)

type queryInput struct {
	expr      expression.Expression
	limit     int
	IndexName string
	order     orderType
	// enfore the query returns exactly one item or err
	singleItem bool
	// can the results of the query be nil
	canBeNil bool
	// return the count of items (as len of list)
	countOnly bool
}

func (table *DdbTable) queryGSI0(ctx context.Context, q queryInput) (Items, error) {
	q.IndexName = GSI0IndexName
	return table.query(ctx, q)
}

func (table *DdbTable) queryGSI1(ctx context.Context, q queryInput) (Items, error) {
	q.IndexName = GSI1IndexName
	return table.query(ctx, q)
}

func (table *DdbTable) queryGSI2(ctx context.Context, q queryInput) (Items, error) {
	q.IndexName = GSI2IndexName
	return table.query(ctx, q)
}

func (table *DdbTable) queryGSI3(ctx context.Context, q queryInput) (Items, error) {
	q.IndexName = GSI1IndexName
	return table.query(ctx, q)
}

func (table *DdbTable) query(ctx context.Context, q queryInput) (Items, error) {
	var indexName *string
	if q.IndexName != "" {
		indexName = aws.String(q.IndexName)
	}

	var limit *int64
	if q.limit > 0 {
		limit = aws.Int64(int64(q.limit))
	} else if q.singleItem {
		limit = aws.Int64(1)
	}

	var order *bool
	switch q.order {
	case orderAsc:
		order = aws.Bool(true)
	case orderDesc:
		order = aws.Bool(false)
	}
	attributes := dynamodb.SelectAllAttributes
	if q.countOnly {
		attributes = dynamodb.SelectCount
	} else if indexName != nil {
		attributes = dynamodb.SelectAllProjectedAttributes
	}

	input := &dynamodb.QueryInput{
		IndexName:                 indexName,
		Limit:                     limit,
		ScanIndexForward:          order,
		TableName:                 aws.String(table.Name),
		KeyConditionExpression:    q.expr.KeyCondition(),
		ExpressionAttributeNames:  q.expr.Names(),
		ExpressionAttributeValues: q.expr.Values(),
		FilterExpression:          q.expr.Filter(),
		Select:                    &attributes,
	}
	inputHash, items := Get(ctx, input.GoString())
	if items != nil {
		return items, nil
	}
	var output *dynamodb.QueryOutput
	var err error

	queryAgain := true
	for queryAgain {
		output, err = table.ddb.QueryWithContext(ctx, input)
		if err != nil {
			return nil, err
		}
		// if selecting only the count, append a empty list
		if q.countOnly {
			items = append(items, make([]map[string]*dynamodb.AttributeValue, *output.Count)...)
		} else {
			items = append(items, output.Items...)
		}
		// done querying when LastEvaluatedKey is nil or we have reached the limit
		queryAgain = nil != output.LastEvaluatedKey
		if output.LastEvaluatedKey == nil {
			queryAgain = false
		} else if q.limit > 0 && len(items) == q.limit {
			queryAgain = false
		}
		input.ExclusiveStartKey = output.LastEvaluatedKey
	}
	// validate that only a single item returned
	err = func() error {
		if !q.singleItem {
			return nil
		}
		// zero items are allowed to exist if canBeNil
		if len(items) == 0 && q.canBeNil {
			return nil
		}

		// zero items are allowed to exist if countOnly
		if len(items) == 0 && q.countOnly {
			return nil
		}

		// else must return a single item
		if len(items) != 1 {
			return fmt.Errorf("wrong number of results reutrned")
		}
		return nil
	}()
	if err != nil {
		return nil, err
	}

	// dont cache count queries because its just... hard
	if !q.countOnly {
		Add(ctx, inputHash, items)
	}
	return items, nil
}

func (table *DdbTable) queryUpdates(ctx context.Context, updateType string, startTime time.Time) (Items, error) {
	keyBuilder := expression.KeyEqual(expression.Key(UpdatesPK), expression.Value(updateType)).
		And(expression.KeyGreaterThanEqual(expression.Key(UpdatesSK), expression.Value(startTime.String())))
	built, err := expression.NewBuilder().WithKeyCondition(keyBuilder).Build()
	if err != nil {
		return nil, err
	}

	return table.query(ctx, queryInput{
		expr:      built,
		IndexName: UpdatesIndexName,
	})
}

// build a update expression on a entry
// return its new version id if the entry is versioned
func buildUpdateExpression(entry DdbEntry) (string, expression.Expression) {
	builder := expression.NewBuilder()
	// enforce the swap on the current value
	if version, ok := entry.(Version); ok {
		// update the new value
		v := version.Version()
		if v == "" {
			v = "x"
		}
		builder = builder.WithCondition(expression.Equal(expression.Name(VersionAttribute), expression.Value(v)))
	}
	update := marshalEntry(entry)

	updateExpression := expression.UpdateBuilder{}
	for k, v := range update {
		if k != "PK" && k != "SK" {
			updateExpression = updateExpression.Set(expression.Name(k), expression.Value(v))
		}
	}

	expr, err := builder.WithUpdate(updateExpression).Build()
	if err != nil {
		panic(err)
	}
	if update.Version() {
		return update.VersionValue(), expr
	}

	return "", expr
}

func (table *DdbTable) UpdateEntries(ctx context.Context, entries ...DdbEntry) ([]string, error) {
	count := len(entries) / 25
	newVersions := make([]string, len(entries))
	for i := 0; i < count; i++ {
		start := (i * count) * 25
		end := start + 25
		if end > len(entries) {
			end = len(entries)
		}
		newV, err := table.Transact(ctx, nil, nil, entries[start:end])
		if err != nil {
			return nil, err
		}
		newVersions = append(newVersions, newV...)
	}
	return newVersions, nil
}

func (table *DdbTable) UpdateEntry(ctx context.Context, entry DdbEntry) (string, error) {
	// create a swap

	//build, err := expression.NewBuilder().WithCondition()
	newVersion, updateExpr := buildUpdateExpression(entry)

	_, err := table.ddb.UpdateItemWithContext(ctx, &dynamodb.UpdateItemInput{
		Key:                       keyFromEntry(entry),
		TableName:                 aws.String(table.Name),
		UpdateExpression:          updateExpr.Update(),
		ExpressionAttributeValues: updateExpr.Values(),
		ExpressionAttributeNames:  updateExpr.Names(),
		ConditionExpression:       updateExpr.Condition(),
	})

	if err != nil {
		return "", err
	}
	Bust(ctx, entry)

	return newVersion, nil
}

func (table *DdbTable) DeleteWithMatchingSkPkBeginsWith(ctx context.Context, sk, pk string) error {
	items, err := table.queryGSI0(ctx, queryInput{
		expr: getBySkPkStartsWith(sk, pk),
	})
	if err != nil {
		return err
	}
	return table.DeleteAll(ctx, unmarhsalIntoBasic(items))
}

func (table *DdbTable) DeleteWithMatchingPK(ctx context.Context, pk string) error {

	items, err := table.query(ctx, queryInput{
		expr: getAllByPK(pk),
	})
	if err != nil {
		return err
	}
	return table.DeleteAll(ctx, unmarhsalIntoBasic(items))
}

func (table *DdbTable) DeleteAll(ctx context.Context, deletes []DdbEntry) error {
	count := int(math.Ceil(float64(len(deletes)) / 25.0))
	for i := 0; i < count; i++ {
		start := (i * count) * 25
		end := start + 25
		if end > len(deletes) {
			end = len(deletes)
		}
		_, err := table.Transact(ctx, nil, nil, deletes[start:end])
		if err != nil {
			return err
		}
	}
	return nil
}

// transact a database
// return any new versions or error
func (table *DdbTable) Transact(ctx context.Context, newEntries, updates []DdbEntry, deletes []DdbEntry) ([]string, error) {

	transtItems := make([]*dynamodb.TransactWriteItem, len(newEntries)+len(updates)+len(deletes))
	newEntryExpression, err := expression.NewBuilder().WithCondition(
		expression.AttributeNotExists(expression.Name(PK)).
			And(expression.AttributeNotExists(expression.Name(SK)))).
		Build()
	if err != nil {
		return nil, err
	}
	versions := make([]string, len(newEntries)+len(updates))

	for i, newEntry := range newEntries {
		entry := marshalEntry(newEntry)
		transtItems[i] = &dynamodb.TransactWriteItem{
			Put: &dynamodb.Put{
				TableName:                 aws.String(table.Name),
				ExpressionAttributeNames:  newEntryExpression.Names(),
				ExpressionAttributeValues: newEntryExpression.Values(),
				ConditionExpression:       newEntryExpression.Condition(),
				Item:                      entry,
			},
		}

		if entry.Version() {
			versions[i] = entry.VersionValue()
		}
	}

	startIndex := len(newEntries)
	for i, update := range updates {
		newVersion, updateExpr := buildUpdateExpression(update)
		transtItems[startIndex+i] = &dynamodb.TransactWriteItem{
			Update: &dynamodb.Update{
				Key:                       keyFromEntry(update),
				TableName:                 aws.String(table.Name),
				UpdateExpression:          updateExpr.Update(),
				ConditionExpression:       updateExpr.Condition(),
				ExpressionAttributeValues: updateExpr.Values(),
				ExpressionAttributeNames:  updateExpr.Names(),
			},
		}
		versions[startIndex+i] = newVersion
	}
	startIndex += len(updates)
	for i, d := range deletes {
		transtItems[startIndex+i] = &dynamodb.TransactWriteItem{
			Delete: &dynamodb.Delete{
				TableName: aws.String(table.Name),
				Key:       keyFromEntry(d),
			},
		}
	}

	_, err = table.ddb.TransactWriteItemsWithContext(ctx, &dynamodb.TransactWriteItemsInput{
		TransactItems: transtItems,
	})
	if err != nil {
		return nil, err
	}
	for _, item := range updates {
		Bust(ctx, item)
	}

	for _, item := range deletes {
		Bust(ctx, item)
	}

	return versions, nil
}

func keyFromEntry(entry DdbEntry) map[string]*dynamodb.AttributeValue {
	return map[string]*dynamodb.AttributeValue{
		PK: {
			S: entry.PK(),
		},
		SK: {
			S: entry.SK(),
		},
	}

}

// create a new entry in ddb
func (table *DdbTable) newEntry(ctx context.Context, entry DdbEntry) error {
	// the PK and and SK must not exist
	expr, err := expression.NewBuilder().WithCondition(
		expression.AttributeNotExists(expression.Name(PK)).
			And(expression.AttributeNotExists(expression.Name(SK)))).
		Build()
	if err != nil {
		return err
	}
	newEntry := marshalEntry(entry)
	// the first version of each object is a x
	if _, ok := newEntry[VersionAttribute]; ok {
		newEntry[VersionAttribute].S = aws.String("x")
	}
	input := &dynamodb.PutItemInput{
		TableName:                 aws.String(table.Name),
		ConditionExpression:       expr.Condition(),
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		Item:                      newEntry,
	}
	_, err = table.ddb.PutItemWithContext(ctx, input)
	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok {
			switch awsErr.Code() {
			case dynamodb.ErrCodeConditionalCheckFailedException:
				return KeyAlreadyExists
			}
		}
		return err
	}
	return nil
}

func (table *DdbTable) pkSkExist(ctx context.Context, pk, sk string) (bool, error) {
	resp, err := table.ddb.GetItemWithContext(ctx, &dynamodb.GetItemInput{
		TableName:            aws.String(table.Name),
		Key:                  keyFromEntry(basicEntry{pk, sk}),
		ProjectionExpression: aws.String(fmt.Sprintf("%s,%s", PK, SK)),
	})
	if err != nil {
		return false, err
	}
	return resp.Item != nil, nil
}

// load a value with the SK=s and keys that begin with sk
// used for GSI0 Queries
func getBySkPkStartsWith(sk, pk string) expression.Expression {
	expr, err := expression.NewBuilder().WithKeyCondition(
		expression.KeyEqual(expression.Key(SK), expression.Value(sk)).And(expression.KeyBeginsWith(expression.Key(PK), pk))).Build()
	if err != nil {
		panic(err)
	}
	return expr
}

// used for GSI0 Queries
// load a value by its pk and sk
func getBySkPk(sk, pk string) expression.Expression {
	expr, err := expression.NewBuilder().WithKeyCondition(
		expression.KeyEqual(expression.Key(PK), expression.Value(pk)).And(expression.KeyEqual(expression.Key(SK), expression.Value(sk)))).Build()
	if err != nil {
		panic(err)
	}
	return expr
}

// load a value with the PK=pk and keys that begin with sk
func getByPkSkStartsWith(pk, sk string) expression.Expression {
	expr, err := expression.NewBuilder().WithKeyCondition(
		expression.KeyEqual(expression.Key(PK), expression.Value(pk)).And(expression.KeyBeginsWith(expression.Key(SK), sk))).Build()
	if err != nil {
		panic(err)
	}
	return expr
}

func getByPkSkBetween(pk, skStart, skEnd string) expression.Expression {
	expr, err := expression.NewBuilder().WithKeyCondition(
		expression.KeyEqual(expression.Key(PK), expression.Value(pk)).And(expression.KeyBetween(expression.Key(SK), expression.Value(skStart), expression.Value(skEnd)))).Build()
	if err != nil {
		panic(err)
	}
	return expr
}

func getByPkSkGraterThanEqual(pk, sk string) expression.Expression {
	expr, err := expression.NewBuilder().WithKeyCondition(
		expression.KeyEqual(expression.Key(PK), expression.Value(pk)).And(expression.KeyGreaterThanEqual(expression.Key(SK), expression.Value(sk)))).Build()
	if err != nil {
		panic(err)
	}
	return expr
}

func getByGSI1PkSkEqual(gsi1Pk, gsi1Sk string) expression.Expression {
	expr, err := expression.NewBuilder().WithKeyCondition(
		expression.KeyEqual(expression.Key(GSI1PK), expression.Value(gsi1Pk)).And(expression.KeyEqual(expression.Key(GSI1SK), expression.Value(gsi1Sk)))).Build()
	if err != nil {
		panic(err)
	}
	return expr
}

func getByGSI1PkSkBetween(gsi1Pk, gsi1SkStart, gsi1SkEnd string) expression.Expression {
	expr, err := expression.NewBuilder().WithKeyCondition(
		expression.KeyEqual(expression.Key(GSI1PK), expression.Value(gsi1Pk)).And(expression.KeyBetween(expression.Key(GSI1SK), expression.Value(gsi1SkStart), expression.Value(gsi1SkEnd)))).Build()
	if err != nil {
		panic(err)
	}
	return expr
}

func getAllByPK(pk string) expression.Expression {
	expr, err := expression.NewBuilder().WithKeyCondition(
		expression.KeyEqual(expression.Key(PK), expression.Value(pk))).Build()
	if err != nil {
		panic(err)
	}
	return expr
}

// load a value by its pk and sk
func getByPkSk(pk, sk string) expression.Expression {
	expr, err := expression.NewBuilder().WithKeyCondition(
		expression.KeyEqual(expression.Key(PK), expression.Value(pk)).And(expression.KeyEqual(expression.Key(SK), expression.Value(sk)))).Build()
	if err != nil {
		panic(err)
	}
	return expr
}

func getByGSI1Pk(gsi1PK string) expression.Expression {
	expr, err := expression.NewBuilder().WithKeyCondition(
		expression.KeyEqual(expression.Key(GSI1PK), expression.Value(gsi1PK))).Build()
	if err != nil {
		panic(err)
	}
	return expr
}

func getByGSI2Pk(gsiPK string) expression.Expression {
	expr, err := expression.NewBuilder().WithKeyCondition(
		expression.KeyEqual(expression.Key(GSI2PK), expression.Value(gsiPK))).Build()
	if err != nil {
		panic(err)
	}
	return expr
}

func getByGSI3Pk(gsiPK string) expression.Expression {
	expr, err := expression.NewBuilder().WithKeyCondition(
		expression.KeyEqual(expression.Key(GSI2PK), expression.Value(gsiPK))).Build()
	if err != nil {
		panic(err)
	}
	return expr
}

func unmarhsalIntoBasic(items Items) []DdbEntry {
	entries := make([]DdbEntry, len(items))
	for i, item := range items {
		b := basicEntry{}
		unmarshalEntry(item, &b)
		entries[i] = b
	}
	return entries
}
