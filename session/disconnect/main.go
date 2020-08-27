package main

import (
	"context"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/jonsabados/pointypoints/api"
	"github.com/jonsabados/pointypoints/lambdautil"
	"github.com/jonsabados/pointypoints/lock"
	"github.com/jonsabados/pointypoints/logging"
	"github.com/jonsabados/pointypoints/session"
	"github.com/rs/zerolog"
)

func NewHandler(prepareLogs logging.Preparer, disconnect session.Disconnector) func(ctx context.Context, request events.APIGatewayWebsocketProxyRequest) (events.APIGatewayProxyResponse, error) {
	return func(ctx context.Context, request events.APIGatewayWebsocketProxyRequest) (events.APIGatewayProxyResponse, error) {
		ctx = prepareLogs(ctx)
		zerolog.Ctx(ctx).Info().Interface("request", request).Msg("disconnect called")
		err := disconnect(ctx, request.RequestContext.ConnectionID)
		if err != nil {
			zerolog.Ctx(ctx).Error().Str("error", fmt.Sprintf("%+v", err)).Msg("error disconnecting user")
			return api.NewInternalServerError(ctx), nil
		}
		return api.NewSuccessResponse(ctx, "OK"), nil
	}
}

func main() {
	lambdautil.CoreStartup()

	logPreparer := logging.NewPreparer()
	sess := lambdautil.DefaultAWSConfig()

	dynamo := lambdautil.NewDynamoClient(sess)
	loader := session.NewLoader(dynamo, lambdautil.SessionTable)
	locker := lock.NewGlobalLockAppropriator(dynamo, lambdautil.LockTable, lambdautil.LockWaitTime, lambdautil.LockExpiration)
	notifier := session.NewChangeNotifier(dynamo, lambdautil.WatcherTable, lambdautil.NewProdMessageDispatcher())
	saveSess := session.NewSaver(dynamo, lambdautil.SessionTable, notifier, lambdautil.SessionTimeout)
	disconnect := session.NewDisconnector(dynamo, lambdautil.InterestTable, locker, loader, saveSess)

	lambda.Start(NewHandler(logPreparer, disconnect))
}
