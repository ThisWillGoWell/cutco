package starketext

import (
	"flag"
	"go.uber.org/zap"
	"os"
)

// create a Logger for a lambda function base, this allows all children loggers produced
//// log these fields. Called on Lambda Startup
//func NewLambdaBaseLogger() (context.Context, *zap.SugaredLogger) {
//	return NewLogger(context.Background(),
//		"lambda_function", os.Getenv("AWS_LAMBDA_FUNCTION_NAME"),
//		"lambda_version", os.Getenv("AWS_LAMBDA_FUNCTION_VERSION"))
//}

//// attach the aws request id and other lambda information for a lambda invoke
//func NewLambdaInvokeLogger(ctx context.Context, with ...interface{}) context.Context {
//	with = append(with, []interface{}{"lambda_function", os.Getenv("AWS_LAMBDA_FUNCTION_NAME"),
//		"lambda_version", os.Getenv("AWS_LAMBDA_FUNCTION_VERSION")}...)
//	lambdaCtx, ok := lambdacontext.FromContext(ctx)
//	if !ok {
//		ctx, _ = NewLogger(ctx, with...)
//		return ctx
//	}
//	with = append(with, "lambda_request_id")
//	with = append(with, lambdaCtx.AwsRequestID)
//	ctx, _ = NewLogger(ctx, with...)
//	return ctx
//}

// create or attach to a zap Logger stored on the context
// this allows for each function to write in a log that has context
// with things attached to it by its parents, or just create a fresh Logger
// if its the first one in its chain to call this
func newLogger(logger *zap.SugaredLogger, with ...interface{}) *zap.SugaredLogger {
	var log *zap.Logger
	if logger == nil {
		var err error
		// if we are in a gotest, no logs unless env var set
		if flag.Lookup("test.v") != nil && os.Getenv("LOG") == "" {
			log = zap.NewNop()
		} else {

			// use production logs because tbh they are easier to read
			log, err = zap.NewProduction()
		}
		if err != nil {
			panic(err)
		}
		logger = log.Sugar()
	}

	// apply any fields if any
	logger = logger.With(with...)
	// add the Logger onto the context of the request
	return logger
}
