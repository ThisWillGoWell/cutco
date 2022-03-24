package storage

import (
	"context"
	"fmt"
	"stock-simulator-serverless/src/models"
)

const OwnerPrefix = models.OwnerPrefix
const CompanyPrefix = models.CompanyPrefix
const SharePrefix = models.SharePrefix

type companyTable struct {
	*DdbTable
}

func (table *companyTable) CreateCompany(ctx context.Context, company *models.CompanyStruct) error {
	// attempt to create a Company with a random id,
	// 10 tries in case the id it picks is already picked
	var err error
	for i := 0; i < 10; i++ {
		company.ID = models.NewUUID()
		err = table.newEntry(ctx, company)
		if err == nil {
			return nil
		}
		if err != KeyAlreadyExists {
			return err
		}
	}
	return err
}

func (table *companyTable) CompanySymbolAvailable(ctx context.Context, symbol string) (bool, error) {
	items, err := table.queryGSI2(ctx, queryInput{
		expr:       getByGSI2Pk(CompanyPrefix + symbol),
		countOnly:  true,
		singleItem: true,
	})
	if err != nil {
		return false, err
	}
	if len(items) == 0 {
		return true, nil
	}
	return false, nil
}

func (table *companyTable) CompanyNameAvailable(ctx context.Context, symbol string) (bool, error) {
	items, err := table.queryGSI2(ctx, queryInput{
		expr:       getByGSI3Pk(CompanyPrefix + symbol),
		countOnly:  true,
		singleItem: true,
	})
	if err != nil {
		return false, err
	}
	if len(items) == 0 {
		return true, nil
	}
	return false, nil
}

//
//func (table *companyTable) LoadCompany(ctx context.Context, request models.ReadCompaniesRequest) (models.CompanyStruct, error) {
//	company, err := table.loadCompany(ctx, request)
//	if err != nil {
//		return models.CompanyStruct{}, err
//	}
//	return company, nil
//}

func (table *companyTable) LoadCompany(ctx context.Context, request models.ReadCompaniesRequest) (*models.CompanyStruct, error) {
	if request.OwnerID == "" && request.CompanyID == "" {
		return nil, fmt.Errorf("invalud ReadCompany request")
	}
	results, err := table.LoadCompanies(ctx, request)
	if err != nil {
		return nil, err
	}
	if len(results) == 0 {
		return nil, NoEntriesFound
	}
	return results[0], nil
}

//func (table *companyTable) LoadCompanies(ctx context.Context, request models.ReadCompaniesRequest) ([]*models.CompanyStruct, error) {
//	companies, err := table.loadCompanies(ctx, request)
//	if err != nil {
//		return nil, err
//	}
//	return models.CompanyListFuncToStruct(companies), nil
//}

func (table *companyTable) LoadCompanies(ctx context.Context, request models.ReadCompaniesRequest) ([]*models.CompanyStruct, error) {
	var companies []*models.CompanyStruct

	if request.CompanyID != "" {
		request.CompanyIds = []string{request.CompanyID}
	}

	// load all Company ids using GSI0
	// or load companyID by OwnerID
	if request.OwnerID == "" && request.CompanyIds == nil {
		items, err := table.queryGSI0(ctx, queryInput{
			expr: getBySkPkStartsWith(CompanyPrefix, CompanyPrefix),
		})
		if err != nil {
			return nil, err
		}
		companies = unmarshalCompanies(items)
	} else if request.OwnerID != "" {
		// fetch Company by ownerID
		items, err := table.queryGSI1(ctx, queryInput{singleItem: true, expr: getByGSI1Pk(OwnerPrefix + request.OwnerID)})
		if err != nil {
			return nil, err
		}
		companies = unmarshalCompanies(items)
	} else if request.CompanyIds != nil {
		companies = make([]*models.CompanyStruct, len(request.CompanyIds))
		for i, id := range request.CompanyIds {
			companies[i] = &models.CompanyStruct{ID: id}
		}
	}
	// now that we have a list of companies, query
	// load all information for the companies
	if request.Selects.SelectInfo || request.Selects.Users != nil {
		entries := make([]DdbEntry, len(companies))
		for i, c := range companies {
			entries[i] = c
		}
		items, err := table.getItems(ctx, entries)
		if err != nil {
			return nil, err
		}
		companies = unmarshalCompanies(items)
	}

	// load share info for companies
	if request.Selects.Shares != nil {
		for i := range companies {
			var err error
			companies[i].Shares, err = table.Company.LoadShares(ctx, models.ReadSharesRequest{
				RequestingID: request.RequestingID,
				CompanyID:    companies[i].ID,
				Selects:      *request.Selects.Shares,
			})
			if err != nil {
				return nil, err
			}
		}
	}
	// load user info
	if request.Selects.Users != nil {
		for i := range companies {
			var err error
			if companies[i].Owner == nil {
				continue
			}
			companies[i].Owner, err = table.Users.LoadUser(ctx, models.ReadUsersRequest{
				RequestingID: request.RequestingID,
				UserID:       companies[i].Owner.ID,
				Selects:      *request.Selects.Users,
			})
			if err != nil {
				return nil, err
			}
		}
	}

	if request.Selects.Transaction != nil {
		if request.CompanyID != "" {
			transactions, err := table.Company.LoadTransactions(ctx, models.ReadTransactionsRequest{
				RequestingID: request.RequestingID,
				CompanyID:    request.CompanyID,
				Selects:      *request.Selects.Transaction,
			})
			if err != nil {
				return nil, err
			}
			for _, c := range companies {
				c.Transactions = transactions
			}
		} else {
			// request the history of each share
			// if HolderID is defined, it will only get transactions for that holder
			for _, c := range companies {
				var err error
				c.Transactions, err = table.LoadTransactions(ctx, models.ReadTransactionsRequest{
					RequestingID: request.RequestingID,
					HolderID:     request.OwnerID,
					CompanyID:    c.ID,
					Selects:      *request.Selects.Transaction,
				})
				if err != nil {
					return nil, err
				}
			}
		}
	}

	return companies, nil

}
func unmarshalCompanies(items Items) []*models.CompanyStruct {
	companies := make([]*models.CompanyStruct, len(items))
	for i, item := range items {
		companies[i] = &models.CompanyStruct{}
		unmarshalEntry(item, companies[i])
	}
	return companies
}
