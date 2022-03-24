package seed

import (
	"stock-simulator-serverless/src/graph/model"
	"stock-simulator-serverless/src/models"
)

func UserStructTSignup(user models.PrivateUserStruct) model.SignupInput {
	return model.SignupInput{
		Login:        user.Login,
		DisplayName:  user.User.Name,
		Password:     user.Password,
		InvestorType: user.User.InvestorType,
		Company: &model.CreateCompanyInput{
			Name:        user.User.Company.Name,
			Symbol:      user.User.Company.Symbol,
			Description: user.User.Description,
			Type:        user.User.Company.Type,
		},
	}
}
