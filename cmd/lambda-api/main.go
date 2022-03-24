package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"stock-simulator-serverless/cmd"
)

var base *cmd.Base

func init() {
	base = cmd.Start()
}

func Handler(ctx context.Context, req events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	r, _ := json.Marshal(req)
	fmt.Println(string(r))
	code, resp := base.Execute(ctx, req.Headers["authorization"], req.Body)
	return events.APIGatewayV2HTTPResponse{
		StatusCode: code,
		Body: resp,
		IsBase64Encoded: false,
	}, nil
}

func main() {
	lambda.Start(Handler)
}
