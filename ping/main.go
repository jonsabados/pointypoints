package main

import (
	"context"
	"fmt"

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
			Body: "pong",
		})
		if err != nil {
			zerolog.Ctx(ctx).Error().Str("error", fmt.Sprintf("%+v", err)).Msg("error dispatching message")
		}
		return api.NewSuccessResponse(ctx, "pong"), nil
	}
}

func main() {
	lambdautil.CoreStartup()
	logPreparer := logging.NewPreparer()
	lambda.Start(NewHandler(logPreparer, lambdautil.NewProdMessageDispatcher()))
}