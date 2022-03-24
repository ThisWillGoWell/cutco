package storage

import (
	"context"
	"fmt"
	"go.uber.org/zap"
	"stock-simulator-serverless/src/errors"
	"stock-simulator-serverless/src/models"
	"stock-simulator-serverless/src/starketext"
)

type usersTable struct {
	*DdbTable
}

func (table *usersTable) CreateNewUser(ctx context.Context, userPrivate *models.PrivateUserStruct) error {

	// attempt to create a user with a random id,
	// 10 tries in case the id it picks is already picked
	var err error
	for i := 0; i < 10; i++ {
		userPrivate.User.ID = models.NewUUID()
		_, err := table.Transact(ctx, []DdbEntry{userPrivate, userPrivate.User}, nil, nil)
		if err == nil {
			return nil
		}
		if err != KeyAlreadyExists {
			return err
		}
	}
	userPrivate.User.VersionID = "x"
	return err
}

//func (table *usersTable) LoadUser(ctx context.Context, request models.ReadUsersRequest) (models.UserStruct, error) {
//	user, err := table.Users.loadUser(ctx, request)
//	if err != nil {
//		return models.UserStruct{}, err
//	}
//	return user(), nil
//}
//
//func (table *usersTable) LoadUsers(ctx context.Context, request models.ReadUsersRequest) ([]models.UserStruct, error) {
//	users, err := table.Users.loadUsers(ctx, request)
//	if err != nil {
//		return nil, err
//	}
//	return models.UserListFuncToStruct(users), nil
//}

func (table *usersTable) LoginAvailable(ctx context.Context, login string) (bool, error) {
	items, err := table.queryGSI1(ctx, queryInput{
		expr:       getByGSI1Pk(UserIDPrefix + login),
		countOnly:  true,
		singleItem: true,
	})
	if err != nil {
		return false, err
	}
	return len(items) == 0, nil

}

func (table *usersTable) LoadPrivateUser(ctx context.Context, request models.ReadPrivateUsers) (*models.PrivateUserStruct, error) {
	//
	var userPrivateItem Item
	var userPrivateItems Items
	var err error
	if request.UserID != "" {
		userPrivateItem, err = table.getItem(ctx, models.PrivateUserStruct{User: &models.UserStruct{ID: request.UserID}})
	} else if request.Login != "" {
		userPrivateItems, err = table.queryGSI1(ctx, queryInput{
			expr:       getByGSI1Pk(UserIDPrefix + request.Login),
			singleItem: true,
		})
		if err == nil {
			userPrivateItem = userPrivateItems[0]
		}
	} else if request.Email != "" {
		userPrivateItems, err = table.queryGSI2(ctx, queryInput{
			expr:       getByGSI2Pk(UserIDPrefix + request.Email),
			singleItem: true,
		})
		if err != nil {
			userPrivateItem = userPrivateItems[0]
		}
	}
	if err != nil {
		return nil, err
	}
	if userPrivateItem == nil {
		return nil, errors.InvalidInputError("login", "login not found")
	}
	userPrivate := &models.PrivateUserStruct{}
	unmarshalEntry(userPrivateItem, userPrivate)

	if request.Selects.User != nil {
		userPrivate.User, err = table.Users.LoadUser(ctx, models.ReadUsersRequest{
			RequestingID: request.RequestingID,
			UserID:       userPrivate.User.ID,
			Selects:      *request.Selects.User,
		})
		if err != nil {
			return nil, err
		}
	}
	return userPrivate, nil
}

func (table *usersTable) LoadUser(ctx context.Context, request models.ReadUsersRequest) (*models.UserStruct, error) {
	if request.UserID == "" {
		return nil, fmt.Errorf("invalid loadUser request")
	}
	if request.UserID != "" {
		request.UserIDs = []string{request.UserID}
	}

	val, err := table.LoadUsers(ctx, request)
	if err != nil {
		return nil, err
	}
	if len(val) == 0 {
		return nil, NoEntriesFound
	}
	return val[0], nil
}

func (table *usersTable) LoadUsers(ctx context.Context, request models.ReadUsersRequest) ([]*models.UserStruct, error) {
	var users []*models.UserStruct
	logger := starketext.LocalLogger(ctx)

	if request.UserIDs != nil && len(request.UserIDs) == 0 {
		logger.Error("read request missing required values")
		return nil, fmt.Errorf("missing user lookup values")
	} else if request.UserIDs != nil {
		logger.Debugw("load users by user-ids")
		users = make([]*models.UserStruct, len(request.UserIDs))
		for i, id := range request.UserIDs {
			users[i] = &models.UserStruct{ID: id}
		}
	} else if request.UserIDs == nil {
		logger.Debugw("load all users")
		items, err := table.queryGSI0(ctx, queryInput{
			expr: getBySkPkStartsWith(UserIDPrefix, UserIDPrefix),
		})
		if err != nil {
			logger.Errorw("failed to load users", zap.Error(err))
			return nil, err
		}
		users = unmarshalUsers(items)
	}
	// remove requesting id
	if request.IgnoreRequestingID {
		for i, u := range users {
			if u.ID == request.RequestingID {
				users = append(users[0:i], users[i+1:]...)
				break
			}
		}
	}

	// load rich information
	if request.Selects.SelectInfo || request.ValidateIDs {
		logger.Debug("loading rich information")
		entries := make([]DdbEntry, len(users))
		for i, user := range users {
			entries[i] = user
		}
		items, err := table.getItems(ctx, entries)
		if err != nil {
			logger.Errorw("failed to load users", zap.Error(err))
			return nil, err
		}
		users = unmarshalUsers(items)
	}

	// load Company
	if request.Selects.Company != nil {
		for i, u := range users {
			var err error
			users[i].Company, err = table.Company.LoadCompany(ctx, models.ReadCompaniesRequest{
				RequestingID: request.RequestingID,
				OwnerID:      u.ID,
				Selects:      *request.Selects.Company,
			})
			if err != nil {
				return nil, err
			}
		}
	}
	// load shares for Users
	if request.Selects.Share != nil {
		for i, u := range users {
			var err error
			users[i].Shares, err = table.Company.LoadShares(ctx, models.ReadSharesRequest{
				Selects:      *request.Selects.Share,
				RequestingID: request.RequestingID,
				HolderID:     u.ID,
			})
			if err != nil {
				return nil, err
			}
		}
	}

	return users, nil

}

func unmarshalUsers(items Items) []*models.UserStruct {
	users := make([]*models.UserStruct, len(items))
	for i, item := range items {
		users[i] = &models.UserStruct{}
		unmarshalEntry(item, users[i])
	}
	return users
}

func unmarshalUser(item Item) models.UserStruct {
	u := models.UserStruct{}
	unmarshalEntry(item, &u)
	return u
}
