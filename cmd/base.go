package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/99designs/gqlgen/graphql"
	"github.com/99designs/gqlgen/graphql/executor"
	"go.uber.org/zap"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"stock-simulator-serverless/src/errors"
	"stock-simulator-serverless/src/graph"
	"stock-simulator-serverless/src/graph/generated"
	"stock-simulator-serverless/src/logic"
	"stock-simulator-serverless/src/seed"
	"stock-simulator-serverless/src/starketext"
	"stock-simulator-serverless/src/storage"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
)

type Base struct {
	Logic   *logic.Logic
	Storage *storage.DdbTable
	exec    graphql.GraphExecutor
}

func StartLocal(seedFn seed.Seed) {
	var b *Base
	if os.Getenv("ITEM_TABLE_NAME") != "" {
		b = Start()
	} else {
		b = &Base{
			Storage: storage.NewLocalDdb(),
		}
		b.Logic = logic.New(b.Storage)
		schema := generated.NewExecutableSchema(generated.Config{Resolvers: &graph.Resolver{Logic: b.Logic}})
		b.exec = executor.New(schema)
		if seedFn != nil {
			seedFn(b.Logic, b.Storage)
		}
	}

	http.HandleFunc("/graph", func(writer http.ResponseWriter, request *http.Request) {
		bodyBytes, err := ioutil.ReadAll(request.Body)
		if err != nil {
			writer.WriteHeader(500)
			_, _ = writer.Write([]byte("failed to read body: " + err.Error()))
			return
		}
		code, resp := b.Execute(request.Context(), request.Header.Get("Authorization"), string(bodyBytes))
		writer.WriteHeader(code)
		_, err = writer.Write([]byte(resp))
		if err != nil {
			fmt.Println("failed to write response err=" + err.Error())
		}
	})
	go func() {
		_ = http.ListenAndServe(":8080", nil)
	}()
}

func Start() *Base {
	sess := session.Must(session.NewSession())
	ddb := dynamodb.New(sess)
	b := &Base{
		Storage: storage.New(os.Getenv("ITEM_TABLE_NAME"), ddb),
	}
	b.Logic = logic.New(b.Storage)
	schema := generated.NewExecutableSchema(generated.Config{Resolvers: &graph.Resolver{Logic: b.Logic}})
	b.exec = executor.New(schema)
	return b
}

func jsonDecode(r io.Reader, val interface{}) error {
	dec := json.NewDecoder(r)
	dec.UseNumber()
	return dec.Decode(val)
}

func (b *Base) Execute(ctx context.Context, token string, query string) (code int, responseBody string) {
	ctx = graphql.StartOperationTrace(ctx)
	ctx, c := starketext.ExtractCtx(ctx)

	if token != "" {
		// authenticate the request
		t, err := b.Storage.Token.LoadToken(ctx, token)
		if errors.Equal(err, errors.NoEntriesFound) {
			c.Logger.Infow("token not found")
			return http.StatusForbidden, "missing token"
		} else if err != nil {
			c.Logger.Error("failed to load token", zap.Error(err))
			return http.StatusInternalServerError, "something bad happened"
		}

		c.AuthedUserID = &t.User.ID
		c.Logger = c.Logger.With("user_id", c.AuthedUserID)
		c.Logger.Infow("authenticated")
	}

	var params *graphql.RawParams
	start := graphql.Now()
	if err := jsonDecode(bytes.NewBuffer([]byte(query)), &params); err != nil {
		return 500, "something bad happened"
	}
	params.ReadTime = graphql.TraceTiming{
		Start: start,
		End:   graphql.Now(),
	}
	rc, gqlErr := b.exec.CreateOperationContext(ctx, params)

	if gqlErr != nil {
		resp := b.exec.DispatchError(graphql.WithOperationContext(ctx, rc), gqlErr)
		respString, err := json.Marshal(resp)
		if err != nil {
			return 500, "something bad happened"
		}
		return 200, string(respString)
	}
	rc.DisableIntrospection = false

	ctx = graphql.WithOperationContext(ctx, rc)
	responses, ctx := b.exec.DispatchOperation(ctx, rc)
	respString, err := json.Marshal(responses(ctx))
	if err != nil {
		return 500, "something bad happened"
	}
	return 200, string(respString)
}
