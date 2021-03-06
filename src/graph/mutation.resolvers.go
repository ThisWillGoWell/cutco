package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"fmt"
	"stock-simulator-serverless/src/graph/generated"
	"stock-simulator-serverless/src/graph/model"
)

func (r *mutationResolver) Login(ctx context.Context, input model.LoginInput) (*model.AuthPayload, error) {
	return r.Logic.User.LoginUser(ctx, input)
}

func (r *mutationResolver) Signup(ctx context.Context, input model.SignupInput) (*model.AuthPayload, error) {
	return r.Logic.User.Signup(ctx, input)
}

func (r *mutationResolver) UpdateMe(ctx context.Context, input model.ChangeMeInput) (*model.ChangeMePayload, error) {
	return r.Logic.User.ChangeMe(ctx, input)
}

func (r *mutationResolver) SendChat(ctx context.Context, input model.SendChatInput) (*model.ChatMessage, error) {
	return r.Logic.Chat.SendChat(ctx, input)
}

func (r *mutationResolver) Trade(ctx context.Context, input model.TradeInput) (*model.TradePayload, error) {
	return r.Logic.Company.Trade(ctx, input)
}

// Mutation returns generated.MutationResolver implementation.
func (r *Resolver) Mutation() generated.MutationResolver { return &mutationResolver{r} }

type mutationResolver struct{ *Resolver }

// !!! WARNING !!!
// The code below was going to be deleted when updating resolvers. It has been copied here so you have
// one last chance to move it out of harms way if you want. There are two reasons this happens:
//  - When renaming or deleting a resolver the old code will be put in here. You can safely delete
//    it when you're done.
//  - You have helper methods in this file. Move them out to keep these resolver files clean.
func (r *mutationResolver) DeleteMe(ctx context.Context) (*bool, error) {
	panic(fmt.Errorf("not implemented"))
}
