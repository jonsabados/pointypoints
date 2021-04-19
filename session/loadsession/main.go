package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/rs/zerolog"

	"github.com/jonsabados/pointypoints/api"
	"github.com/jonsabados/pointypoints/lambdautil"
	"github.com/jonsabados/pointypoints/logging"
	"github.com/jonsabados/pointypoints/session"
)

func NewHandler(prepareLogs logging.Preparer, loadSession session.Loader, saveWatcher session.WatcherSaver, dispatch api.MessageDispatcher) func(ctx context.Context, request events.APIGatewayWebsocketProxyRequest) (events.APIGatewayProxyResponse, error) {
	return func(ctx context.Context, request events.APIGatewayWebsocketProxyRequest) (events.APIGatewayProxyResponse, error) {
		ctx = prepareLogs(ctx)
		l := new(session.LoadSessionRequest)
		err := json.Unmarshal([]byte(request.Body), l)
		if err != nil {
			zerolog.Ctx(ctx).Warn().Str("error", fmt.Sprintf("%+v", err)).Msg("error reading load request body")
			return api.NewInternalServerError(ctx, nil), nil
		}
		sess, err := loadSession(ctx, l.SessionID)
		if err != nil {
			zerolog.Ctx(ctx).Error().Str("error", fmt.Sprintf("%+v", err)).Msg("error reading session")
			return api.NewPermissionDeniedResponse(ctx, nil), nil
		}
		if sess == nil {
			zerolog.Ctx(ctx).Warn().Str("sessionID", l.SessionID).Msg("session not found")
			return api.NewPermissionDeniedResponse(ctx, nil), nil
		}
		err = saveWatcher(ctx, sess.SessionID, request.RequestContext.ConnectionID)
		if err != nil {
			zerolog.Ctx(ctx).Error().Str("error", fmt.Sprintf("%+v", err)).Msg("error recording interest")
			return api.NewInternalServerError(ctx, nil), nil
		}
		err = dispatch(ctx, request.RequestContext.ConnectionID, api.Message{
			Type: api.SessionLoaded,
			Body: session.ToParticipantView(*sess, request.RequestContext.ConnectionID),
		})
		if err != nil {
			zerolog.Ctx(ctx).Error().Str("error", fmt.Sprintf("%+v", err)).Msg("error dispatching message")
		}
		return api.NewSuccessResponse(ctx, nil, sess), nil
	}
}

func main() {
	lambdautil.CoreStartup()

	logPreparer := logging.NewPreparer()
	sess := lambdautil.DefaultAWSConfig()

	dynamo := lambdautil.NewDynamoClient(sess)
	loader := session.NewLoader(dynamo, lambdautil.SessionTable)
	dispatcher := lambdautil.NewProdMessageDispatcher()
	watcherSaver := session.NewWatcherSaver(dynamo, lambdautil.SessionTable, lambdautil.SessionTimeout)

	lambda.Start(NewHandler(logPreparer, loader, watcherSaver, dispatcher))
}
