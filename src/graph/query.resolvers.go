package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"fmt"
	"stock-simulator-serverless/src/graph/generated"
	"stock-simulator-serverless/src/graph/model"
)

func (r *queryResolver) Me(ctx context.Context) (*model.MeUser, error) {
	return r.Logic.User.Me(ctx)
}

func (r *queryResolver) User(ctx context.Context, input model.GetUsersInput) (*model.User, error) {
	if input.UserID == "" {
		return nil, fmt.Errorf("missing userdao id")
	}
	users, err := r.Users(ctx, &input)
	if err != nil {
		return nil, err
	}
	if len(users) == 0 {
		return nil, nil
	}
	return users[0], nil
}

func (r *queryResolver) Users(ctx context.Context, input *model.GetUsersInput) ([]*model.User, error) {
	return r.Logic.User.Users(ctx, input)
}

func (r *queryResolver) Company(ctx context.Context, input model.GetCompanyInput) (*model.Company, error) {
	if input.CompanyID == nil {
		return nil, fmt.Errorf("missing companyID")
	}

	companies, err := r.Companies(ctx, &input)
	if err != nil {
		return nil, err
	}
	if len(companies) == 0 {
		return nil, nil
	}
	return companies[0], nil
}

func (r *queryResolver) Companies(ctx context.Context, input *model.GetCompanyInput) ([]*model.Company, error) {
	return r.Logic.Company.GetCompanies(ctx, input)
}

func (r *queryResolver) Share(ctx context.Context, input *model.GetShareInput) (*model.Share, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *queryResolver) Chat(ctx context.Context, input *model.ReadChatMessagesInput) ([]*model.ChatChannel, error) {
	return r.Logic.Chat.Chats(ctx, input)
}

// Query returns generated.QueryResolver implementation.
func (r *Resolver) Query() generated.QueryResolver { return &queryResolver{r} }

type queryResolver struct{ *Resolver }
