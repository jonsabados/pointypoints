package main

import (
	"context"
	"fmt"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/rs/zerolog"

	"github.com/jonsabados/pointypoints/lambdautil"
	"github.com/jonsabados/pointypoints/logging"
	"github.com/jonsabados/pointypoints/session"
)

func NewHandler(prepareLogs logging.Preparer, disconnect session.Disconnector) func(ctx context.Context, request events.APIGatewayWebsocketProxyRequest) (events.APIGatewayProxyResponse, error) {
	return func(ctx context.Context, request events.APIGatewayWebsocketProxyRequest) (events.APIGatewayProxyResponse, error) {
		ctx = prepareLogs(ctx)
		zerolog.Ctx(ctx).Info().Interface("request", request).Msg("disconnect called")
		err := disconnect(ctx, request.RequestContext.ConnectionID)
		if err != nil {
			zerolog.Ctx(ctx).Error().Str("error", fmt.Sprintf("%+v", err)).Msg("error disconnecting user")
			return events.APIGatewayProxyResponse{
				StatusCode: http.StatusInternalServerError,
			}, nil
		}
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusNoContent,
		}, nil
	}
}

func main() {
	lambdautil.CoreStartup()

	logPreparer := logging.NewPreparer()
	sess := lambdautil.DefaultAWSConfig()

	dynamo := lambdautil.NewDynamoClient(sess)
	loader := session.NewLoader(dynamo, lambdautil.SessionTable)
	notifier := session.NewChangeNotifier(dynamo, lambdautil.SessionTable, lambdautil.NewProdMessageDispatcher())
	disconnect := session.NewDisconnector(dynamo, lambdautil.SessionTable, lambdautil.SessionSocketIndex, loader, notifier)

	lambda.Start(NewHandler(logPreparer, disconnect))
}
