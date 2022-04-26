package logic

import (
	"context"
	"fmt"
	"stock-simulator-serverless/src/errors"
	"stock-simulator-serverless/src/graph/model"
	"stock-simulator-serverless/src/models"
	"stock-simulator-serverless/src/selection"
	"stock-simulator-serverless/src/starketext"
	"stock-simulator-serverless/src/storage"
	"time"
)

type companyLogic struct {
	*Logic
}

func (c *companyLogic) validate(ctx context.Context, input *models.CompanyStruct) error {
	if err := validateCompany(input); err != nil {
		return err
	}

	// verify the name is not taken
	ok, err := c.storage.Company.CompanyNameAvailable(ctx, input.Name)
	if err != nil {
		return err
	}
	if !ok {
		return errors.InvalidInputError("company name", "already taken")
	}

	// verify the symbol is open
	ok, err = c.storage.Company.CompanySymbolAvailable(ctx, input.Symbol)
	if err != nil {
		return err
	}
	if !ok {
		return errors.InvalidInputError("symbol", "already taken")
	}
	return nil
}

func (c *companyLogic) CreateCompany(ctx context.Context, input models.CompanyStruct, validate bool) (*models.CompanyStruct, error) {
	if validate {
		if err := c.validate(ctx, &input); err != nil {
			return nil, err
		}
	}
	// give the company some defaults
	if input.Value == 0 {
		input.Value = 1000

	}
	if input.OpenShares == 0 {
		input.OpenShares = 100
	}

	input.CreatedAt = time.Now()
	if err := c.storage.Company.CreateCompany(ctx, &input); err != nil {
		return nil, err
	}
	return &input, nil
}

func (c *companyLogic) GetCompanies(ctx context.Context, input *model.GetCompanyInput) ([]*model.Company, error) {
	companies, err := c.GetCompaniesStruct(ctx, selection.CompanySelects(ctx), input)
	if err != nil {
		return nil, err
	}
	return companyStructsToGql(companies), nil
}

func (c *companyLogic) GetCompaniesStruct(ctx context.Context, selects selection.Company, input *model.GetCompanyInput) ([]*models.CompanyStruct, error) {

	request := models.ReadCompaniesRequest{
		Selects: selects,
	}
	if input != nil {
		if input.CompanyID != nil {
			request.CompanyID = *input.CompanyID
		} else {
			request.CompanyIds = input.CompanyIDs
		}
	}

	// load userID for request
	userID, ok := starketext.AuthenticatedID(ctx)
	if !ok {
		return nil, errors.MissingAuthentication
	}

	request.RequestingID = userID

	// load a single company
	companies, err := c.storage.Company.LoadCompanies(ctx, request)
	if err != nil {
		return nil, err
	}

	return companies, nil
}

func (c *companyLogic) Trade(ctx context.Context, input model.TradeInput) (*model.TradePayload, error) {
	// load the userdao
	user, err := c.LoadAuthedUser(ctx)
	if err != nil {
		return nil, err
	}
	trade := models.Trade{
		User:  user,
		Count: int(input.Amount),
	}

	trade.User = user
	// company does not have the value hydraded at this point, so hydrade it
	trade.Company, err = c.storage.Company.LoadCompany(ctx, models.ReadCompaniesRequest{
		RequestingID: user.ID,
		CompanyID:    input.CompanyID,
		Selects: selection.Company{
			SelectInfo: true,
		},
	})
	if err != nil {
		return nil, err
	}

	// is the price correct?
	if trade.Company.Value != int(input.Price) {
		return nil, errors.InvalidInputError("price", "miss-matched price")
	}

	// does a share exist?
	share, err := c.storage.Company.LoadShare(ctx, models.ReadSharesRequest{
		CompanyID: trade.Company.ID,
		HolderID:  trade.User.ID,
		Selects: selection.Share{
			SelectInfo: true,
		},
	})
	noShareExists := err == storage.NoEntriesFound

	if err != nil && err != storage.NoEntriesFound {
		return nil, err
	}

	// buy
	if trade.Count > 0 {
		if trade.Company.OpenShares < trade.Count {
			return nil, fmt.Errorf("not enough open shares!")
		}
		cost := trade.Company.Value * trade.Count

		if trade.User.Wallet < cost {
			return nil, fmt.Errorf("not enough money")
		}
		trade.User.Wallet -= cost
		trade.Company.OpenShares -= trade.Count

	} else {
		if noShareExists {
			return nil, fmt.Errorf("you dont own any of those shares")
		}
		// sell
		if share.Count < -1*trade.Count {
			return nil, fmt.Errorf("you dont own enough of those shares")
		}
		trade.User.Wallet += trade.Count * trade.Company.Value
		trade.Company.OpenShares += trade.Count
	}
	// create a transaction history item
	transaction := models.TransactionStruct{
		Company: trade.Company,
		User:    user,
		Time:    time.Now(),
		Value:   trade.Company.Value,
		Count:   trade.Count,
	}

	if noShareExists {
		newShare := models.ShareStruct{
			Company: trade.Company,
			Holder:  trade.User,
			Count:   trade.Count,
		}

		_, err = c.storage.Transact(ctx,
			[]storage.DdbEntry{newShare, transaction},
			[]storage.DdbEntry{trade.User, trade.Company},
			nil,
		)
		if err != nil {
			return nil, err
		}
	} else {
		share.Count += trade.Count
		_, err = c.storage.Transact(ctx, []storage.DdbEntry{transaction}, []storage.DdbEntry{share, trade.User, trade.Company}, nil)
		return nil, err
	}
	return nil, nil

}

func (c *companyLogic) CreateShare(ctx context.Context, input models.ShareStruct) error {
	if input.Company == nil {
		return fmt.Errorf("must provide a company")
	}
	if input.Holder == nil {
		return fmt.Errorf("must provide holder")
	}
	return c.storage.Company.CreateShare(ctx, input)
}
