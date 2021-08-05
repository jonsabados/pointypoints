package main

import (
	"context"
	"encoding/json"
	"os"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/jonsabados/goauth/aws"
	"github.com/rs/zerolog"

	"github.com/jonsabados/pointypoints/api"
	"github.com/jonsabados/pointypoints/cors"
	"github.com/jonsabados/pointypoints/lambdautil"
	"github.com/jonsabados/pointypoints/logging"
	"github.com/jonsabados/pointypoints/profile"
	"github.com/jonsabados/pointypoints/session"
)

func NewHandler(prepareLogs logging.Preparer, corsHeaders cors.ResponseHeaderBuilder, loadSession session.Loader, saveWatcher session.WatcherSaver, dispatch api.MessageDispatcher) func(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	return func(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
		ctx = prepareLogs(ctx)
		sessionID := request.PathParameters["session"]

		w := new(session.WatchSessionRequest)
		err := json.Unmarshal([]byte(request.Body), w)
		if err != nil {
			zerolog.Ctx(ctx).Warn().Err(err).Msg("error reading watch session request body")
			return api.NewInternalServerError(ctx, corsHeaders(ctx, request.Headers)), nil
		}

		sess, err := loadSession(ctx, sessionID)
		if err != nil {
			zerolog.Ctx(ctx).Error().Err(err).Msg("error reading session")
			return api.NewPermissionDeniedResponse(ctx, corsHeaders(ctx, request.Headers)), nil
		}
		if sess == nil {
			zerolog.Ctx(ctx).Warn().Str("sessionID", sessionID).Msg("session not found")
			return api.NewPermissionDeniedResponse(ctx, nil), nil
		}

		principal, err := aws.ExtractPrincipal(request)
		if err != nil {
			zerolog.Ctx(ctx).Warn().Err(err).Msg("error extracting principal")
			return api.NewInternalServerError(ctx, corsHeaders(ctx, request.Headers)), nil
		}


		err = saveWatcher(ctx, principal, sess.SessionID, w.ConnectionID)
		if err != nil {
			zerolog.Ctx(ctx).Error().Err(err).Msg("error recording interest")
			return api.NewInternalServerError(ctx, corsHeaders(ctx, request.Headers)), nil
		}
		err = dispatch(ctx, w.ConnectionID, api.Message{
			Type: api.SessionUpdated,
			Body: session.ToParticipantView(*sess, w.ConnectionID),
		})
		if err != nil {
			zerolog.Ctx(ctx).Error().Err(err).Msg("error dispatching message")
		}
		return api.NewNoContentResponse(ctx, corsHeaders(ctx, request.Headers)), nil
	}
}

func main() {
	lambdautil.CoreStartup()

	logPreparer := logging.NewPreparer()
	sess := lambdautil.DefaultAWSConfig()

	statsFactory := profile.NewStatsUpdateFactory(lambdautil.ProfileTable)

	dynamo := lambdautil.NewDynamoClient(sess)
	loader := session.NewLoader(dynamo, lambdautil.SessionTable)
	dispatcher := lambdautil.NewProdMessageDispatcher()
	watcherSaver := session.NewWatcherSaver(dynamo, lambdautil.SessionTable, lambdautil.SessionTimeout, statsFactory)

	allowedDomains := strings.Split(os.Getenv("ALLOWED_ORIGINS"), ",")

	lambda.Start(NewHandler(logPreparer, cors.NewResponseHeaderBuilder(allowedDomains), loader, watcherSaver, dispatcher))
}
