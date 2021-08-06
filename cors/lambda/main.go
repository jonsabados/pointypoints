package main

import (
	"context"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-xray-sdk-go/xray"

	"github.com/jonsabados/pointypoints/cors"
	"github.com/jonsabados/pointypoints/lambdautil"
	"github.com/jonsabados/pointypoints/logging"
)

func newHandler(prepareLogs logging.Preparer, headers cors.ResponseHeaderBuilder) func (ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	return func(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
		ctx = prepareLogs(ctx)

		return events.APIGatewayProxyResponse{
			StatusCode:        http.StatusNoContent,
			Headers:           headers(ctx, request.Headers),
			Body:              "",
			IsBase64Encoded:   false,
		}, nil
	}
}


func main() {
	lambdautil.CoreStartup()
	logPreparer := logging.NewPreparer()

	err := xray.Configure(xray.Config{
		LogLevel: "warn",
	})
	if err != nil {
		panic(err)
	}

	allowedDomains := lambdautil.AllowedCORSOrigins()
	lambda.Start(newHandler(logPreparer, cors.NewResponseHeaderBuilder(allowedDomains)))
}