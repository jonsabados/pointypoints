package main

import (
	"context"
	"net/http"
	"os"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-xray-sdk-go/xray"

	"github.com/jonsabados/pointypoints/cors"
)

func newHandler(headers cors.ResponseHeaderBuilder) func (ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	return func(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
		return events.APIGatewayProxyResponse{
			StatusCode:        http.StatusNoContent,
			Headers:           headers(request.Headers),
			Body:              "",
			IsBase64Encoded:   false,
		}, nil
	}
}


func main() {
	err := xray.Configure(xray.Config{
		LogLevel: "warn",
	})
	if err != nil {
		panic(err)
	}

	allowedDomains := strings.Split(os.Getenv("ALLOWED_ORIGINS"), ",")
	lambda.Start(newHandler(cors.NewResponseHeaderBuilder(allowedDomains)))
}