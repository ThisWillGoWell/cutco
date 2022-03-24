package logic

import (
	gql "stock-simulator-serverless/src/graph/model"
	"stock-simulator-serverless/src/models"
	"time"
)

func signupToUser(input gql.SignupInput) *models.PrivateUserStruct {

	user := &models.UserStruct{
		Name:         input.DisplayName,
		Description:  input.Description,
		InvestorType: input.InvestorType,
	}

	if input.Company != nil {
		user.Company = &models.CompanyStruct{
			Name:        input.Company.Name,
			CreatedAt:   time.Now(),
			Symbol:      input.Company.Symbol,
			Description: input.Company.Description,
			Owner:       user,
		}

	}
	e := ""
	if input.Email != nil {
		e = *input.Email
	}

	return &models.PrivateUserStruct{
		Email:    e,
		User:     user,
		Login:    input.Login,
		Password: input.Password,
	}
}

func listOfUserStruct(input []*models.UserStruct) []*gql.User {
	if input == nil {
		return nil
	}
	output := make([]*gql.User, len(input))
	for i, user := range input {
		output[i] = userToGql(user)
	}
	return output
}

func listOfUser(input []*models.UserStruct) []*gql.User {
	if input == nil {
		return nil
	}
	output := make([]*gql.User, len(input))
	for i, user := range input {
		output[i] = userToGql(user)
	}
	return output
}

func privateUserToGql(user *models.PrivateUserStruct) *gql.MeUser {
	if user == nil {
		return nil
	}
	var email *string
	if user.Email != "" {
		email = &user.Email
	}
	gqlUser := &gql.MeUser{
		User:  userToGql(user.User),
		Login: user.Login,
		Email: email,
	}
	return gqlUser
}

func userToGql(user *models.UserStruct) *gql.User {
	if user == nil {
		return nil
	}
	gqlUser := &gql.User{
		ID:           user.ID,
		Name:         user.Name,
		CreatedAt:    user.CreatedAt,
		LastActiveAt: user.LastActiveAt,
		Company:      companyToGql(user.Company),
		Shares:       sharesToGql(user.Shares),
		Description:  user.Description,
		InvestorType: user.InvestorType,
	}
	return gqlUser
}

func sharesToGql(input []*models.ShareStruct) []*gql.Share {
	if input == nil {
		return nil
	}
	output := make([]*gql.Share, len(input))
	for i, item := range input {
		output[i] = shareToGql(item)
	}
	return output
}

func shareToGql(share *models.ShareStruct) *gql.Share {
	if share == nil {
		return nil
	}
	return &gql.Share{
		Count:        share.Count,
		Company:      companyToGql(share.Company),
		Holder:       userToGql(share.Holder),
		Transactions: transactionsToGql(share.Transactions),
	}
}

func companyStructsToGql(input []*models.CompanyStruct) []*gql.Company {
	if input == nil {
		return nil
	}
	output := make([]*gql.Company, len(input))
	for i, item := range input {
		output[i] = companyToGql(item)
	}
	return output
}

func transactionsToGql(input []*models.TransactionStruct) []*gql.Transaction {
	if input == nil {
		return nil
	}
	output := make([]*gql.Transaction, len(input))
	for i, item := range input {
		output[i] = transactionToGql(item)
	}
	return output
}

func transactionToGql(transaction *models.TransactionStruct) *gql.Transaction {
	if transaction == nil {
		return nil
	}
	return &gql.Transaction{
		Count: transaction.Count,
		Time:  transaction.Time,
		Value: transaction.Value,
		User:  userToGql(transaction.User),
	}
}

func companyToGql(company *models.CompanyStruct) *gql.Company {
	if company == nil {
		return nil
	}

	return &gql.Company{
		ID:           company.ID,
		Name:         company.Name,
		Owner:        userToGql(company.Owner),
		CreatedAt:    company.CreatedAt,
		Symbol:       company.Symbol,
		Description:  company.Description,
		Shares:       sharesToGql(company.Shares),
		Value:        company.Value,
		Transactions: transactionsToGql(company.Transactions),
	}
}

func listOfMessages(input []*models.ChatMessage) []*gql.ChatMessage {
	if input == nil {
		return nil
	}
	output := make([]*gql.ChatMessage, len(input))
	for i, m := range input {
		output[i] = messageToGql(m)
	}
	return output
}

func messageToGql(message *models.ChatMessage) *gql.ChatMessage {
	return &gql.ChatMessage{
		ID:            message.ID,
		User:          userToGql(message.Owner),
		Channel:       channelToGql(message.Channel),
		Message:       message.Message,
		CreatedAt:     message.CreatedAt,
		PaginationKey: models.TimeStringValue(message.CreatedAt),
	}
}

func channelToGql(channel *models.ChatChannelStruct) *gql.ChatChannel {
	if channel == nil {
		return nil
	}
	result := &gql.ChatChannel{
		ID:        channel.ID,
		Name:      &channel.Name,
		CreatedAt: &channel.CreatedAt,
		Members:   listOfUser(channel.Members),
		Messages:  listOfMessages(channel.Messages),
	}
	switch channel.Type {
	case models.ChatChannelTypePublic:
		result.Type = gql.ChannelTypePublic
	case models.ChatChannelTypeWhisper:
		result.Type = gql.ChannelTypeWhisper
	case models.ChatChannelTypePrivate:
		result.Type = gql.ChannelTypePrivate
	}
	return result
}

func listOfChannels(input []*models.ChatChannelStruct) []*gql.ChatChannel {
	if input == nil {
		return nil
	}
	output := make([]*gql.ChatChannel, len(input))
	for i, item := range input {
		output[i] = channelToGql(item)
	}
	return output
}
