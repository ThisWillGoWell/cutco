package logic

import (
	"context"
	"go.uber.org/zap"
	"stock-simulator-serverless/src/models"
	"stock-simulator-serverless/src/selection"
	"stock-simulator-serverless/src/starketext"
	"stock-simulator-serverless/src/storage"
)

func (l *Logic) deleteCompany(ctx context.Context, c models.CompanyStruct) error {
	logger := starketext.LocalLogger(ctx, "company_id", c.ID)
	// return all the users $$$
	comp, err := l.storage.Company.LoadCompany(ctx, models.ReadCompaniesRequest{
		CompanyID: c.ID,
		Selects: selection.Company{
			Shares: &selection.Share{
				SelectInfo: true,
				Holder: &selection.User{
					SelectInfo: true,
				},
				Transaction: &selection.Transaction{},
			},
		},
	})
	if err != nil {
		logger.Errorw("failed to load company for delete", zap.Error(err))
		return err
	}
	c = *comp

	// set the available stocks to 0 so no one can buy any more
	c.OpenShares = 0
	c.VersionID, err = l.storage.UpdateEntry(ctx, c)
	if err != nil {
		logger.Errorw("failed to set the shares to zero during a delete", zap.Error(err))
	}
	// sell and delete the stocks
	logger.Infow("selling shares", "count", len(c.Shares))
	for _, s := range c.Shares {
		// add the value of the stock directly to the owner
		s.Holder.Wallet += c.Value * s.Count
		_, err = l.storage.UpdateEntry(ctx, s.Holder)
		if err != nil {
			logger.Errorw("failed to update user during sell", "user_id", s.Holder.ID, zap.Error(err))
			return err
		}

		// delete the stock and all the entries
		deletes := make([]storage.DdbEntry, len(s.Transactions)+1)
		deletes[0] = s
		for i, trans := range s.Transactions {
			deletes[i+1] = trans
		}

		if err = l.storage.DeleteAll(ctx, deletes); err != nil {
			logger.Error("failed to delete shares and transactions", "user_id", s.Holder.ID, zap.Error(err))
			return err
		}
		logger.Infow("deleted stock", "user_id", s.Holder.ID)
	}
	// delete anything else under the company prefix
	return l.storage.Company.DeleteWithMatchingPK(ctx, *c.PK())

}

func (l *Logic) deleteAccount(ctx context.Context, u *models.PrivateUserStruct) error {
	logger := starketext.LocalLogger(ctx)
	logger.Infow("deleting messages")
	// delete messages
	err := l.storage.Chat.DeleteMessagesByUserID(ctx, u.User.ID)
	if err != nil {
		return err
	}

	// delete company if the user is CEO of one
	company, err := l.storage.Company.LoadCompany(ctx, models.ReadCompaniesRequest{
		OwnerID:      u.User.ID,
		RequestingID: u.User.ID,
		Selects:      selection.Company{},
	})
	if err != nil {
		return err
	}

	if company != nil {
		logger.Infow("deleting company the user owns")
		if err := l.deleteCompany(ctx, *company); err != nil {
			logger.Errorw("failed to delete company", zap.Error(err))
			return err
		}
	}

	// delete all tokens for the user
	err = l.storage.DeleteWithMatchingSkPkBeginsWith(ctx, *u.PK(), models.TokenPrefix)

	// return all shares the user owned
	shares, err := l.storage.Company.LoadShares(ctx, models.ReadSharesRequest{
		RequestingID: u.User.ID,
		HolderID:     u.User.ID,
		Selects: selection.Share{
			SelectInfo: true,
			Company: &selection.Company{
				SelectInfo: true,
			},
		},
	})
	if err != nil {
		logger.Errorw("failed to load shares during user delete", zap.Error(err))
		return err
	}
	for _, s := range shares {
		s.Company.OpenShares += s.Count
		_, err = l.storage.Transact(ctx, nil, []storage.DdbEntry{s.Company}, []storage.DdbEntry{s})
		if err != nil {
			logger.Errorw("failed to delete share and update company", "company_id", company.ID, zap.Error(err))
			return err
		}
	}

	// delete the user and anything under their pk
	err = l.storage.DeleteWithMatchingPK(ctx, *u.PK())
	if err != nil {
		logger.Errorw("failed to delete user", zap.Error(err))
		return err
	}
	logger.Infow("user deleted")
	return nil

}
