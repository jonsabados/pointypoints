package main

import (
	"context"
	"fmt"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/rs/zerolog"

	"github.com/jonsabados/pointypoints/api"
	"github.com/jonsabados/pointypoints/lambdautil"
	"github.com/jonsabados/pointypoints/logging"
)

func NewHandler(prepareLogs logging.Preparer, dispatch api.MessageDispatcher) func(ctx context.Context, request events.APIGatewayWebsocketProxyRequest) (events.APIGatewayProxyResponse, error) {
	return func(ctx context.Context, request events.APIGatewayWebsocketProxyRequest) (events.APIGatewayProxyResponse, error) {
		ctx = prepareLogs(ctx)

		err := dispatch(ctx, request.RequestContext.ConnectionID, api.Message{
			Type: api.Ping,
			Body: struct {
				Message      string `json:"message"`
				ConnectionID string `json:"connectionId"`
			}{
				Message:      "pong",
				ConnectionID: request.RequestContext.ConnectionID,
			},
		})
		if err != nil {
			zerolog.Ctx(ctx).Error().Str("error", fmt.Sprintf("%+v", err)).Msg("error dispatching message")
		}
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusNoContent,
		}, nil
	}
}

func main() {
	lambdautil.CoreStartup()
	logPreparer := logging.NewPreparer()
	lambda.Start(NewHandler(logPreparer, lambdautil.NewProdMessageDispatcher()))
}
