package models

import (
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
)

const CompanyPrefix = "company#"
const OwnerPrefix = "owner#"
const ValuePrefix = "value#"
const SharePrefix = "share#"
const TransactionPrefix = "transaction#"

type CompanyStruct struct {
	ID           string                `json:"-"`
	Name         string                `json:"-"`
	CreatedAt    time.Time             `json:"created_at"`
	Symbol       string                `json:"-"`
	Description  string                `json:"description"`
	Type         string                `json:"type"`
	OpenShares   int                   `json:"-"`
	Value        int                   `json:"-"`
	Owner        *UserStruct           `json:"-"`
	ValueHistory []CompanyValueHistory `json:"-"`
	Shares       []*ShareStruct        `json:"-"`
	Transactions []*TransactionStruct  `json:"-"`
	VersionID    string                `json:"-"`
}

func (c CompanyStruct) Version() string {
	return c.VersionID
}

func (c *CompanyStruct) VersionLoad(s string) {
	c.VersionID = s
}

func (CompanyStruct) Json() {}

// info is stored under CompanyStruct#ID / CompanyStruct#
func (c CompanyStruct) PK() *string {
	return aws.String(CompanyPrefix + c.ID)
}

func (c CompanyStruct) SK() *string {
	return aws.String(CompanyPrefix)
}

func (c *CompanyStruct) Load(pk, sk string) {
	c.ID = RemovePrefix(pk)
}

//GSI1 Stores OwnerID -> Company
func (c CompanyStruct) GSI1PK() *string {
	if c.Owner == nil {
		return nil
	}
	return aws.String(OwnerPrefix + c.Owner.ID)
}
func (c CompanyStruct) GSI1SK() *string {
	if c.Owner == nil {
		return nil
	}
	return aws.String(OwnerPrefix)
}
func (c *CompanyStruct) GSI1Load(pk, _ string) {
	c.Owner = &UserStruct{
		ID: RemovePrefix(pk),
	}
}

//GSI2 Stores Name -> x
func (c CompanyStruct) GSI2PK() *string {
	return aws.String(CompanyPrefix + c.Name)
}
func (c CompanyStruct) GSI2SK() *string {
	return aws.String(CompanyPrefix)
}
func (c *CompanyStruct) GSI2Load(pk, _ string) {
	c.Name = RemovePrefix(pk)
}

//GSI3 Stores Symbol -> Company
func (c CompanyStruct) GSI3PK() *string {
	return aws.String(CompanyPrefix + c.Symbol)
}
func (c CompanyStruct) GSI3SK() *string {
	return aws.String(CompanyPrefix)
}
func (c *CompanyStruct) GSI3Load(pk, _ string) {
	c.Symbol = RemovePrefix(pk)
}

// GSI3 StockSymbol -> X
func (c CompanyStruct) Integer() *string {
	return aws.String(fmt.Sprintf("%d", c.Value))
}
func (c *CompanyStruct) IntegerLoad(in int) {
	c.Value = in
}

func (c CompanyStruct) Integer2() *string {
	return aws.String(fmt.Sprintf("%d", c.OpenShares))
}
func (c *CompanyStruct) Integer2Load(in int) {
	c.OpenShares = in
}

// historical timeseries of the price of the CompanyStruct
type CompanyValueHistory struct {
	Value     int       `json:"value"`
	CompanyID string    `json:"-"`
	Time      time.Time `json:"-"`
}

// history is stored under CompanyStruct#ID / value#Time
func (h CompanyValueHistory) PK() *string {
	return aws.String(CompanyPrefix + h.CompanyID)
}

func (h CompanyValueHistory) SK() *string {
	return aws.String(ValuePrefix + TimeStringValue(h.Time))
}

// populate any values possilbe from the pk and sk
func (h CompanyValueHistory) Load(pk, sk string) CompanyValueHistory {
	h.CompanyID = RemovePrefix(pk)
	h.Time = MustParseTimeString(RemovePrefix(sk))
	return h
}
