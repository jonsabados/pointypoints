package main

import (
	"context"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-xray-sdk-go/xray"
	"github.com/jonsabados/pointypoints/api"
	"github.com/jonsabados/pointypoints/logging"
)

func NewHandler(prepareLogs logging.Preparer) func(ctx context.Context, request events.APIGatewayWebsocketProxyRequest) api.Response {
	return func(ctx context.Context, request events.APIGatewayWebsocketProxyRequest) api.Response {
		ctx = prepareLogs(ctx)
		return api.NewSuccessResponse(ctx, "whatever")
	}
}

func main() {
	err := xray.Configure(xray.Config{
		LogLevel: "warn",
	})
	if err != nil {
		panic(err)
	}

	logPreparer := logging.NewPreparer()

	lambda.Start(NewHandler(logPreparer))
}
