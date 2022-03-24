package selection

import (
	"context"
	"reflect"
	"unsafe"

	"github.com/99designs/gqlgen/graphql"
	"github.com/vektah/gqlparser/v2/ast"
)

type ChatChannel struct {
	SelectID       bool `select:"id"`
	SelectInfo     bool `select:"name,created-on"`
	UserSelects    *User
	MessageSelects *ChatMessage
}

type ChatMessage struct {
	UserSelects *User
}

type UserPrivate struct {
	SelectEmail bool
	SelectLogin bool
	SelectInfo  bool
	User        *User
}

type User struct {
	SelectInfo bool
	Share      *Share
	Company    *Company
}

// generic return type of all the different relations that could be found on an object
type selects struct {
	user        *User
	company     *Company
	chatMessage *ChatMessage
	chatChannel *ChatChannel
	share       *Share
}

//// if a key exists in this map, update this boolean to true
//type keyCheck struct {
//	values:
//}
//
//func userSelects(field *ast.Field) {
//
//}
//
//type userSelector struct {
//
//	baseOutput map[string]interface{}
//
//
//}
//
//type selector struct {
//
//}
//
//func selections(field *ast.Field) selects {
//	user := model.User{}
//	b, _ := json.Marshal(user)
//	output := map[string]interface{}{}
//	_ = json.Unmarshal(b, &output)
//
//	for _, selectionSet := range field.SelectionSet {
//		selection := selectionSet.(*ast.Field)
//		if val, ok := output[selection.Name]; ok {
//			switch val.(type) {
//			case *model.User, []*model.User:
//				userSelects(field)
//			case *model.Company, []*model.Company:
//
//			}
//		}
//
//	}
//
//}

func extractResolver(ctx context.Context) *graphql.FieldContext {
	// since we know the top context will be the resovler, we can just take it out
	// we cant read it by key since the key type is private (graphql.key)
	// this is gross but hey

	val := reflect.ValueOf(ctx).Elem()
	rf := val.Field(2)
	rf = reflect.NewAt(rf.Type(), unsafe.Pointer(rf.UnsafeAddr())).Elem()
	if val := rf.Interface(); val != nil {
		fldCtx, ok := val.(*graphql.FieldContext)
		if ok {
			return fldCtx
		}
		return nil
	}
	return nil
}

func ChatChannelSelects(ctx context.Context) ChatChannel {
	resolver := extractResolver(ctx)
	if resolver == nil {
		return ChatChannel{}
	}
	return *chatChannelSelects(resolver.Field.Field)
}

func chatChannelSelects(field *ast.Field) *ChatChannel {
	selects := &ChatChannel{}
	for _, selectionSet := range field.SelectionSet {
		selection := selectionSet.(*ast.Field)
		switch selection.Name {
		case "members":
			selects.UserSelects = userSelections(selection)
		case "messages":
			selects.MessageSelects = chatMessageSelection(selection)
		case "id":
			selects.SelectID = true
		case "type", "created_at", "name":
			selects.SelectInfo = true
		}
	}
	return selects
}

func UserPrivatesSelects(ctx context.Context) UserPrivate {
	resolver := extractResolver(ctx)
	if resolver == nil {
		return UserPrivate{}
	}
	return *userPrivateSelect(resolver.Field.Field)
}

func userPrivateSelect(field *ast.Field) *UserPrivate {
	selects := &UserPrivate{}
	for _, selectionSet := range field.SelectionSet {
		selection := selectionSet.(*ast.Field)
		switch selection.Name {
		case "user":
			selects.User = userSelections(selection)
		case "email":
			selects.SelectEmail = true
		case "login":
			selects.SelectLogin = true
		default:
			selects.SelectInfo = true
		}
	}
	return selects
}

func chatMessageSelection(field *ast.Field) *ChatMessage {
	selects := &ChatMessage{}
	for _, selectionSet := range field.SelectionSet {
		selection := selectionSet.(*ast.Field)
		switch selection.Name {
		case "user":
			selects.UserSelects = userSelections(selection)
		}
	}
	return selects
}

func UserSelects(ctx context.Context) User {
	resolver := extractResolver(ctx)
	if resolver == nil {
		return User{}
	}
	return *userSelections(resolver.Field.Field)
}

func userSelections(field *ast.Field) *User {
	selects := &User{}
	for _, selectionSet := range field.SelectionSet {
		selection := selectionSet.(*ast.Field)
		switch selection.Name {
		case "createdAt", "lastActiveAt", "name", "description":
			selects.SelectInfo = true
		case "shares":
			selects.Share = shareSelections(selection)
		case "company":
			selects.Company = companySelections(selection)
		}
	}
	return selects
}

func shareSelections(field *ast.Field) *Share {
	selects := &Share{}
	for _, selectionSet := range field.SelectionSet {
		selection := selectionSet.(*ast.Field)
		switch selection.Name {
		case "count":
			selects.SelectInfo = true
		case "holder":
			selects.Holder = userSelections(selection)
		case "company":
			selects.Company = companySelections(selection)
		case "transactions":
			selects.Transaction = transactionsSelections(selection)
		}
	}
	return selects
}

func transactionsSelections(field *ast.Field) *Transaction {
	selects := &Transaction{}
	for _, selectionSet := range field.SelectionSet {
		selection := selectionSet.(*ast.Field)
		switch selection.Name {
		case "value", "count", "time":
			selects.SelectInfo = true
		case "user":
			selects.User = userSelections(selection)
		}
	}
	return selects
}

func CompanySelects(ctx context.Context) Company {
	resolver := extractResolver(ctx)
	if resolver == nil {
		return Company{}
	}
	return *companySelections(resolver.Field.Field)
}

func companySelections(field *ast.Field) *Company {
	selects := &Company{}
	for _, selectionSet := range field.SelectionSet {
		selection := selectionSet.(*ast.Field)
		switch selection.Name {
		case "name", "createdAt", "symbol", "description", "value":
			selects.SelectInfo = true
		case "shares":
			selects.Shares = shareSelections(selection)
		case "owner":
			selects.Users = userSelections(selection)
		case "history":
			selects.SelectHistory = true
		case "transactions":
			selects.Transaction = transactionsSelections(selection)
		}
	}
	return selects

}
