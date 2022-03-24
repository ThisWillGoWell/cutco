package logic

import (
	"context"
	"log"
	"stock-simulator-serverless/src/errors"
	"stock-simulator-serverless/src/graph/model"
	"stock-simulator-serverless/src/models"
	"stock-simulator-serverless/src/selection"
	"stock-simulator-serverless/src/starketext"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type userLogic struct {
	*Logic
}

func (logic *userLogic) Me(ctx context.Context) (*model.MeUser, error) {
	// convert the input into a ReadUsersRequest
	userID, ok := starketext.AuthenticatedID(ctx)
	if !ok {
		return nil, errors.MissingAuthentication
	}

	user, err := logic.storage.Users.LoadPrivateUser(ctx, models.ReadPrivateUsers{
		Selects:      selection.UserPrivatesSelects(ctx),
		RequestingID: userID,
		UserID:       userID,
	})
	if err != nil {
		return nil, err
	}

	return privateUserToGql(user), nil
}

func (logic *userLogic) LoginUser(ctx context.Context, input model.LoginInput) (*model.AuthPayload, error) {

	if err := validLogin(input.Login); err != nil {
		return nil, err
	}

	// load user by login
	user, err := logic.storage.Users.LoadPrivateUser(ctx, models.ReadPrivateUsers{
		Login: input.Login,
	})
	if err != nil {
		return nil, err
	}

	// validate password
	if !comparePasswords(user.Password, input.Password) {
		return nil, errors.InvalidInputError("password", "invalid password")
	}

	// create a new token for the user
	token, err := logic.storage.Token.NewToken(ctx, models.Token{
		User:      user.User,
		CreatedAt: time.Now(),
	})
	if err != nil {
		return nil, err
	}

	return &model.AuthPayload{
		Token: token.Token,
	}, nil
}

func (logic *userLogic) Users(ctx context.Context, input *model.GetUsersInput) ([]*model.User, error) {
	users, err := logic.UsersStruct(ctx, selection.UserSelects(ctx), input)
	if err != nil {
		return nil, err
	}
	return listOfUserStruct(users), err
}
func (logic *userLogic) UsersStruct(ctx context.Context, selects selection.User, input *model.GetUsersInput) ([]*models.UserStruct, error) {
	logger := starketext.LocalLogger(ctx)
	logger.Debugw("loading users", "input", input)
	// convert the input into a ReadUsersRequest
	request := models.ReadUsersRequest{
		Selects: selects,
	}
	if input != nil {
		if input.UserID != "" {
			request.UserID = input.UserID
			request.UserIDs = []string{input.UserID}
		} else if input.UserIDs != nil {
			request.UserIDs = input.UserIDs
		} else {
			request.IgnoreRequestingID = true
		}
	} else {
		// fetch all users but me
		request.IgnoreRequestingID = true
	}

	// load userID from token
	var ok bool
	request.RequestingID, ok = starketext.AuthenticatedID(ctx)
	if !ok {
		return nil, errors.MissingAuthentication
	}

	// load the requested info
	users, err := logic.storage.Users.LoadUsers(ctx, request)
	if err != nil {
		return nil, err
	}
	return users, nil
}

func (logic *userLogic) ChangeMe(ctx context.Context, input model.ChangeMeInput) (*model.ChangeMePayload, error) {
	// change password
	if input.Password != nil {
		if input.OldPassword == nil {
			return nil, errors.MissingInputError("OldPassword")
		}
		if err := validPassword(*input.Password); err != nil {
			return nil, err
		}
	}

	// display name change
	if input.DisplayName != nil {
		if err := validDisplayName(*input.DisplayName); err != nil {
			return nil, err
		}
	}

	// description change
	if input.Description != nil {
		if err := validDescription(*input.Description); err != nil {
			return nil, err
		}
	}

	// input validated as best as can be
	// must have valid token
	userID, authed := starketext.AuthenticatedID(ctx)
	if !authed {
		return nil, errors.MissingAuthentication
	}

	// load the private user
	user, err := logic.storage.Users.LoadPrivateUser(ctx, models.ReadPrivateUsers{
		UserID: userID,
		Selects: selection.UserPrivate{
			SelectInfo: true,
			User:       &selection.User{},
		},
	})
	if err != nil {
		return nil, err
	}

	if input.Delete != nil && *input.Delete {
		err = logic.deleteAccount(ctx, user)
		if err != nil {
			return nil, err
		}
		return &model.ChangeMePayload{Success: true}, nil
	}

	// process password change
	if input.Password != nil {
		if !comparePasswords(user.Password, *input.OldPassword) {
			return nil, errors.InvalidInputError("OldPassword", "invalid current password")
		}
		user.Password = HashAndSalt(*input.Password)
	}

	if input.DisplayName != nil {
		user.User.Name = *input.DisplayName
	}

	if input.Description != nil {
		user.User.Description = *input.Description
	}

	// update the entry in the ddb
	_, err = logic.storage.UpdateEntries(ctx, user.User, user)
	if err != nil {
		return nil, err
	}

	return &model.ChangeMePayload{
		Success: true,
		Me:      privateUserToGql(user),
	}, nil
}

// create a new user in the system
func (logic *userLogic) Signup(ctx context.Context, create model.SignupInput) (*model.AuthPayload, error) {
	user := signupToUser(create)
	err := logic.NewUser(ctx, user)
	if err != nil {
		return nil, err
	}
	//create a login token for the user
	token, err := logic.storage.Token.NewToken(ctx, models.Token{User: user.User})
	if err != nil {
		return nil, err
	}
	return &model.AuthPayload{
		Token: token.Token,
	}, nil
}

// NewUser create a new user
func (logic *userLogic) NewUser(ctx context.Context, create *models.PrivateUserStruct) error {
	if err := validateUser(create); err != nil {
		return err
	}

	// Company Validation
	company := create.User.Company
	if company != nil {
		if err := logic.Company.validate(ctx, create.User.Company); err != nil {
			return err
		}
	}

	// make sure login does not exist
	ok, err := logic.storage.Users.LoginAvailable(ctx, create.Login)
	if err != nil {
		return errors.SomethingBadHappened("failed to check if login was available", err)
	}
	if !ok {
		return errors.InvalidInputError("Login", "already taken")
	}

	if create.User.Wallet == 0 {
		create.User.Wallet = 100000
	}
	create.Password = HashAndSalt(create.Password)
	create.User.CreatedAt = time.Now()

	// create a new user
	err = logic.storage.Users.CreateNewUser(ctx, create)
	if err != nil {
		return errors.SomethingBadHappened("failed to make user", err)
	}

	// do we also need to create the company?
	if company != nil {
		company.Owner = create.User
		company.CreatedAt = time.Now()
		company, err = logic.Company.CreateCompany(ctx, *company, false)
		if err != nil {
			return err
		}

		// issue 10,000 shares shares to the owner
		if err := logic.storage.Company.CreateShare(ctx, models.ShareStruct{
			Count:   10_000,
			Company: company,
			Holder:  create.User,
		}); err != nil {
			return err
		}
	}
	return nil

}

func HashAndSalt(password string) string {
	pwd := []byte(password)
	// Use GenerateFromPassword to hash & salt pwd.
	// MinCost is just an integer constant provided by the bcrypt
	// package along with DefaultCost & MaxCost.
	// The cost can be any value you want provided it isn't lower
	// than the MinCost (4)
	hash, err := bcrypt.GenerateFromPassword(pwd, bcrypt.MinCost)
	if err != nil {
		log.Println(err)
	}
	// GenerateFromPassword returns a byte slice so we need to
	// convert the bytes to a string and return it
	return string(hash)
}

func comparePasswords(hashedPwd string, plainPwd string) bool {
	// Since we'll be getting the hashed password from the DB it
	// will be a string so we'll need to convert it to a byte slice
	bytePW := []byte(plainPwd)
	byteHash := []byte(hashedPwd)
	err := bcrypt.CompareHashAndPassword(byteHash, bytePW)
	if err != nil {
		log.Println(err)
		return false
	}

	return true
}
