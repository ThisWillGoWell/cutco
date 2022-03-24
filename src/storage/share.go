package storage

import (
	"context"
	"fmt"
	"stock-simulator-serverless/src/graph/model"
	"stock-simulator-serverless/src/models"
)

func (table *companyTable) CreateShare(ctx context.Context, share models.ShareStruct) error {
	return table.newEntry(ctx, share)
}

func (table *companyTable) LoadShare(ctx context.Context, request models.ReadSharesRequest) (*models.ShareStruct, error) {
	// load by both holder id and by company id
	shares, err := table.LoadShares(ctx, request)
	if err != nil {
		return nil, err
	}
	if shares == nil {
		return nil, NoEntriesFound
	}
	if len(shares) != 1 {
		panic("got wrong number back")
	}
	return shares[0], nil
}

//func (table *companyTable) LoadShares(ctx context.Context, request models.ReadSharesRequest) ([]models.ShareStruct, error) {
//	shares, err := table.loadShares(ctx, request)
//	if err != nil {
//		return nil, err
//	}
//	if shares == nil {
//		return nil, NoEntriesFound
//	}
//	return models.ShareListFuncToStruct(shares), err
//}

func (table *companyTable) LoadShares(ctx context.Context, request models.ReadSharesRequest) ([]*models.ShareStruct, error) {

	var shares []*models.ShareStruct
	shareHydrated := false
	// load all shares ids by owner id using GSI0
	// or load companyID by OwnerID
	if request.HolderID != "" && request.CompanyID != "" {
		// get s single share by company-holder
		items, err := table.queryGSI0(ctx, queryInput{
			expr: getBySkPkStartsWith(SharePrefix+request.HolderID, CompanyPrefix+request.CompanyID),
		})
		if err != nil {
			return nil, err
		}
		shares = unmarshalShares(items)
	} else if request.HolderID != "" {
		items, err := table.queryGSI0(ctx, queryInput{
			expr: getBySkPkStartsWith(SharePrefix+request.HolderID, CompanyPrefix),
		})
		if err != nil {
			return nil, err
		}
		shares = unmarshalShares(items)

	} else if request.CompanyID != "" {
		shareHydrated = true
		items, err := table.query(ctx, queryInput{
			expr: getByPkSkStartsWith(CompanyPrefix+request.CompanyID, SharePrefix),
		})
		if err != nil {
			return nil, err
		}
		shares = unmarshalShares(items)
	} else {
		return nil, fmt.Errorf("invalid read request")
	}
	if shares == nil || len(shares) == 0 {
		return nil, nil
	}

	if request.Selects.SelectInfo && !shareHydrated {

		entries := make([]DdbEntry, len(shares))
		for i, s := range shares {
			entries[i] = s
		}
		items, err := table.getItems(ctx, entries)
		if err != nil {
			return nil, err
		}
		shares = unmarshalShares(items)

	}

	// load share info for shares
	if request.Selects.Company != nil {
		// are we dealing with a single company id
		if request.CompanyID != "" {
			company, err := table.LoadCompany(ctx, models.ReadCompaniesRequest{
				RequestingID: request.RequestingID,
				CompanyID:    request.CompanyID,
				Selects:      *request.Selects.Company,
			})
			if err != nil {
				return nil, err
			}
			for i := range shares {
				shares[i].Company = company
			}
		} else {
			ids := make([]string, len(shares))
			for i, s := range shares {
				ids[i] = s.Company.ID
			}
			// look up all companies for the holder and then match them to the shares
			companies, err := table.LoadCompanies(ctx, models.ReadCompaniesRequest{
				Selects:      *request.Selects.Company,
				RequestingID: request.RequestingID,
				CompanyIds:   ids,
			})
			if err != nil {
				return nil, err
			}
			// match the companies with the shares
			match := make(map[string]int)
			for i := range shares {
				match[shares[i].Company.ID] = i
			}
			for i := range companies {
				pos, ok := match[companies[i].ID]
				if !ok {
					return nil, fmt.Errorf("missing company during match")
				}
				shares[pos].Company = companies[i]
			}

		}
	}
	// load user info
	if request.Selects.Holder != nil {
		for i := range shares {
			var err error

			shares[i].Holder, err = table.Users.LoadUser(ctx, models.ReadUsersRequest{
				RequestingID: request.RequestingID,
				UserID:       shares[i].Holder.ID,
				Selects:      *request.Selects.Holder,
			})
			if err != nil {
				return nil, err
			}
		}
	}

	if request.Selects.Transaction != nil {
		// if we are loading shares for a specific user, only load those transactions
		//if request.HolderID != "" {
		//	transactions, err := table.Company.LoadTransactions(ctx, models.ReadTransactionsRequest{
		//		RequestingID: request.RequestingID,
		//		CompanyID:    request.CompanyID,
		//		Selects:      request.Selects.Transaction,
		//		HolderID:     request.HolderID,
		//	})
		//	if err != nil {
		//		return nil, err
		//	}
		//	for _, s := range shares {
		//		s.Transactions = transactions
		//	}
		//} else {
		// request the history of each share
		// if HolderID is defined, it will only get transactions for that holder
		for _, share := range shares {
			var err error
			share.Transactions, err = table.LoadTransactions(ctx, models.ReadTransactionsRequest{
				RequestingID: request.RequestingID,
				CompanyID:    share.Company.ID,
				HolderID:     share.Holder.ID,
				Selects:      *request.Selects.Transaction,
			})
			if err != nil {
				return nil, err
			}
		}
	}

	return shares, nil
}

func (table *DdbTable) Trade(ctx, input model.TradeInput) {
	//
}

func unmarshalShares(items Items) []*models.ShareStruct {
	shares := make([]*models.ShareStruct, len(items))
	for i, item := range items {
		shares[i] = &models.ShareStruct{}
		unmarshalEntry(item, shares[i])
	}
	return shares
}
