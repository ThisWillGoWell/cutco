package starketext

import (
	"bytes"
	"context"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambdacontext"
	"io/ioutil"
	"net/http"
	"stock-simulator-serverless/src/errors"
	"stock-simulator-serverless/src/models"

	"go.uber.org/zap"
)

// LoadToken define interface for validating tokens
type LoadToken interface {
	LoadToken(ctx context.Context, token string) (models.Token, error)
}

func ApiGatewayMiddleware(loadToken LoadToken) func(ctx context.Context, request events.APIGatewayV2HTTPRequest) *events.APIGatewayV2HTTPResponse {
	return func(ctx context.Context, request events.APIGatewayV2HTTPRequest) *events.APIGatewayV2HTTPResponse {
		ctx, c := ExtractCtx(ctx)
		c.RequestBody = request.Body

		token, ok := request.Headers["Authorization"]
		if ok {
			// authenticate the request
			t, err := loadToken.LoadToken(ctx, token)
			if errors.Equal(err, errors.NoEntriesFound) {
				c.Logger.Infow("token not found")
				return &events.APIGatewayV2HTTPResponse{
					StatusCode: 404,
				}
			} else if err != nil {
				c.Logger.Error("failed to load token", zap.Error(err))
				return  &events.APIGatewayV2HTTPResponse{
					StatusCode: 500,
					Body: "something bad happened",
				}
			}

			c.AuthedUserID = &t.User.ID
			c.Logger = c.Logger.With("user_id", c.AuthedUserID)
			c.Logger.Infow("authenticated")
		}
		return nil
	}
}

func HttpMiddleware(loadToken LoadToken, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// create a new Context for the request
		ctx, c := ExtractCtx(r.Context())
		if r.Body != nil {

			bodyBytes, _ := ioutil.ReadAll(r.Body)
			_ = r.Body.Close()  //  must close
			r.Body = ioutil.NopCloser(bytes.NewBuffer(bodyBytes))

			c.Logger = newLogger(c.Logger, "body", string(bodyBytes))
		}

		// add a requestID uuid
		header := r.Header.Get("Authorization")
		// if auth is not available then proceed to resolver
		if header == "" {
			c.Logger.Infow("unauthenticated request")
			next.ServeHTTP(w, r.WithContext(ctx))
		} else {
			// authenticate the request
			t, err := loadToken.LoadToken(ctx, header)
			if errors.Equal(err, errors.NoEntriesFound) {
				c.Logger.Infow("token not found")
				w.WriteHeader(http.StatusForbidden)
				return
			} else if err != nil {
				c.Logger.Error("failed to load token", zap.Error(err))
				w.WriteHeader(http.StatusInternalServerError)
			}

			c.AuthedUserID = &t.User.ID
			c.Logger = c.Logger.With("user_id", c.AuthedUserID)
			c.Logger.Infow("authenticated")
			next.ServeHTTP(w, r.WithContext(ctx))
		}
	})
}

const CtxKey = "starketext"

// responsible for managing the context in a starket request
type Context struct {
	logic        LoadToken
	requestID    string
	Logger       *zap.SugaredLogger
	RequestBody  string
	AuthedUserID *string
}

// return the current Logger
func Logger(ctx context.Context) *zap.SugaredLogger {
	ctx, c := ExtractCtx(ctx)
	return c.Logger
}

// create a new child Logger of the current Logger, with additional fields
// this overwrites the Logger
func NewLogger(ctx context.Context, with ...interface{}) (context.Context, *zap.SugaredLogger) {
	ctx, c := ExtractCtx(ctx)
	c.Logger = c.Logger.With(with...)
	return ctx, c.Logger
}

// create a new, local Logger with additional fields
func LocalLogger(ctx context.Context, with ...interface{}) *zap.SugaredLogger {
	_, c := ExtractCtx(ctx)
	return c.Logger.With(with...)
}

// create a new ctx that is authenticated for the user id
func NewUserAuthed(userID string) context.Context {
	ctx, c := ExtractCtx(nil)
	c.AuthedUserID = &userID
	return ctx
}

// extract or create a new Context on the context
func ExtractCtx(ctx context.Context) (context.Context, *Context) {
	if ctx == nil {
		ctx = context.Background()
	}
	lc, isLambda := lambdacontext.FromContext(ctx)
	val := ctx.Value(CtxKey)
	var c *Context
	if val == nil {
		if isLambda {
			c = &Context{
				Logger:    newLogger(nil, "request_id", lc.AwsRequestID),
				requestID: lc.AwsRequestID,
			}
		} else {
			requestID := models.NewUUID()
			c = &Context{
				Logger:    newLogger(nil, "request_id", requestID),
				requestID: requestID,
			}
		}

		ctx = context.WithValue(ctx, CtxKey, c)
	} else {
		var ok bool
		c, ok = val.(*Context)
		if !ok {
			panic("wrong type of Context stored in the context")
		}
	}
	return ctx, c
}

// return the id of the user authed on the request
func AuthenticatedID(ctx context.Context) (string, bool) {
	_, c := ExtractCtx(ctx)
	if c.AuthedUserID == nil {
		return "", false
	}
	return *c.AuthedUserID, true
}

func RequestID(ctx context.Context) string {
	_, c := ExtractCtx(ctx)
	return c.requestID
}
