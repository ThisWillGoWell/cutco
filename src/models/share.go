package models

import (
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
)

type Share func() ShareStruct

// A share of ownership in the CompanyStruct
type ShareStruct struct {
	Count        int                  `json:"-"`
	Company      *CompanyStruct       `json:"-"`
	Holder       *UserStruct          `json:"-"`
	Transactions []*TransactionStruct `json:"-"`
	version      string
}

func (s ShareStruct) Version() string {
	return s.version
}

func (s *ShareStruct) VersionLoad(v string) {
	s.version = v
}

// shares is stored under company#ID / share#userID
func (s ShareStruct) PK() *string {
	return aws.String(CompanyPrefix + s.Company.ID)
}

func (s ShareStruct) SK() *string {
	return aws.String(SharePrefix + s.Holder.ID)
}

func (s *ShareStruct) Load(pk, sk string) {
	s.Company = &CompanyStruct{ID: RemovePrefix(pk)}
	s.Holder = &UserStruct{ID: RemovePrefix(sk)}
}

func (s ShareStruct) Integer() *string {
	return aws.String(fmt.Sprintf("%d", s.Count))
}

func (s *ShareStruct) IntegerLoad(in int) {
	s.Count = in
}

// A transaction is a purchase or sale of a share of stock
type TransactionStruct struct {
	Company *CompanyStruct `json:"-"`
	User    *UserStruct    `json:"-"`
	Time    time.Time      `json:"-"`
	Value   int            `json:"value"`
	Count   int            `json:"count"`
}

func (TransactionStruct) Json() {}

// transactions are stored under user#RequestingID / transaction#TimeStamp_ID
// Get all transactions for a Company
func (t TransactionStruct) PK() *string {
	return aws.String(CompanyPrefix + t.Company.ID)
}

func (t TransactionStruct) SK() *string {
	return aws.String(TransactionPrefix + TimeStringValue(t.Time))
}

func (t *TransactionStruct) Load(pk, sk string) {
	t.Company = &CompanyStruct{ID: NthPos(pk, 1)}
	t.Time = MustParseTimeString(RemovePrefix(sk))
}

// Transaction by Company and User
func (t TransactionStruct) GSI1PK() *string {
	return aws.String(CompanyPrefix + t.Company.ID)
}

func (t TransactionStruct) GSI1SK() *string {
	return aws.String(TransactionPrefix + t.User.ID + "#" + TimePrefix + TimeStringValue(t.Time))
}

func (t *TransactionStruct) GSI1Load(_, sk string) {
	// the only thing that is not unique on this index is userid
	t.User = &UserStruct{ID: NthPos(sk, 1)}
}
