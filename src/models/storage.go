package models

import (
	"stock-simulator-serverless/src/selection"
	"time"
)

type ReadChatRequest struct {
	MessageLimit int
	StartTime    time.Time
	ChannelID    string
	RequestingID string
	UserID       string
	Selects      selection.ChatChannel
}

type ReadPrivateUsers struct {
	RequestingID string
	UserID       string

	// login
	Login string

	// emails
	Email string

	// users
	Selects selection.UserPrivate
}

type ReadUsersRequest struct {
	// requesting userid
	RequestingID string
	//Validate the Ids
	ValidateIDs bool

	IgnoreRequestingID bool

	Selects selection.User
	// read many userIDs
	UserIDs []string
	// load just a usedID
	UserID string

	StartTime time.Time
}

type ReadCompaniesRequest struct {
	RequestingID string
	// get by compaines
	CompanyIds []string
	// get by company id
	CompanyID string
	// get company by onwer
	OwnerID string

	Selects selection.Company
}

type ReadSharesRequest struct {
	RequestingID string
	CompanyID    string
	HolderID     string
	Selects      selection.Share
}

type ReadTransactionsRequest struct {
	RequestingID string
	CompanyID    string
	HolderID     string
	StartTime    time.Time
	Limit        int
	Selects      selection.Transaction
}
