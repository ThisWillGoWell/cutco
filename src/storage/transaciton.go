package storage

import (
	"context"
	"fmt"
	"stock-simulator-serverless/src/models"
	"time"
)

const TransactionPrefix = models.TransactionPrefix
const TimePrefix = models.TimePrefix

func (table *companyTable) LoadTransactions(ctx context.Context, request models.ReadTransactionsRequest) ([]*models.TransactionStruct, error) {

	var transactions []*models.TransactionStruct
	transactionHydrated := false
	// load
	if request.CompanyID == "" {
		return nil, fmt.Errorf("invalid transaction read request")
	}

	if request.StartTime.IsZero() {
		request.StartTime = time.Now()
	}

	// load transactions by owner
	if request.HolderID != "" {
		skStart := TransactionPrefix + request.HolderID + TimePrefix
		// get a single share by company-holder
		items, err := table.queryGSI1(ctx, queryInput{
			expr:  getByGSI1PkSkBetween(CompanyPrefix+request.CompanyID, skStart, skStart+models.TimeStringValue(request.StartTime)),
			limit: request.Limit,
			order: orderDesc,
		})
		if err != nil {
			return nil, err
		}
		transactions = unmarshalTransactions(items)
	} else {
		transactionHydrated = true
		items, err := table.query(ctx, queryInput{
			expr:  getByPkSkBetween(CompanyPrefix+request.CompanyID, TransactionPrefix, TransactionPrefix+models.TimeStringValue(request.StartTime)),
			order: orderDesc,
			limit: request.Limit,
		})
		if err != nil {
			return nil, err
		}
		transactions = unmarshalTransactions(items)
	}

	if transactions == nil || len(transactions) == 0 {
		return nil, nil
	}

	if request.Selects.SelectInfo && !transactionHydrated {

		entries := make([]DdbEntry, len(transactions))
		for i, s := range transactions {
			entries[i] = s
		}
		items, err := table.getItems(ctx, entries)
		if err != nil {
			return nil, err
		}
		transactions = unmarshalTransactions(items)
	}

	// load user info
	if request.Selects.User != nil {
		if request.HolderID != "" {
			// load a single user
			user, err := table.Users.LoadUser(ctx, models.ReadUsersRequest{
				RequestingID: request.RequestingID,
				UserID:       request.HolderID,
				Selects:      *request.Selects.User,
			})
			if err != nil {
				return nil, err
			}
			for i := range transactions {
				transactions[i].User = user
			}
		} else {
			// load many users
			// todo batch
			for i := range transactions {
				var err error
				transactions[i].User, err = table.Users.LoadUser(ctx, models.ReadUsersRequest{
					RequestingID: request.RequestingID,
					UserID:       transactions[i].User.ID,
					Selects:      *request.Selects.User,
				})
				if err != nil {
					return nil, err
				}
			}
		}

	}

	return transactions, nil
}

func unmarshalTransactions(items Items) []*models.TransactionStruct {
	shares := make([]*models.TransactionStruct, len(items))
	for i, item := range items {
		shares[i] = &models.TransactionStruct{}
		unmarshalEntry(item, shares[i])
	}
	return shares
}
