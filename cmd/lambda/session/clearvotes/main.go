package main

import (
	"context"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/rs/zerolog"

	"github.com/jonsabados/pointypoints/api"
	"github.com/jonsabados/pointypoints/cors"
	"github.com/jonsabados/pointypoints/lambdautil"
	"github.com/jonsabados/pointypoints/logging"
	"github.com/jonsabados/pointypoints/session"
)

func NewHandler(prepareLogs logging.Preparer, corsHeaders cors.ResponseHeaderBuilder, loadSession session.Loader, saveSession session.Saver) func(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	return func(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
		ctx = prepareLogs(ctx)

		sessionID := request.PathParameters["session"]
		sess, err := loadSession(ctx, request.PathParameters["session"])
		if err != nil {
			zerolog.Ctx(ctx).Error().Err(err).Msg("error reading session")
			return api.NewPermissionDeniedResponse(ctx, corsHeaders(ctx, request.Headers)), nil
		}
		if sess == nil {
			zerolog.Ctx(ctx).Warn().Str("sessionID", sessionID).Msg("session not found")
			return api.NewPermissionDeniedResponse(ctx, corsHeaders(ctx, request.Headers)), nil
		}

		facilitatorKey := api.FacilitatorKey(request.Headers)
		if sess.FacilitatorSessionKey != facilitatorKey {
			zerolog.Ctx(ctx).Warn().Str("sessionID", sessionID).Msg("attempt to show votes with incorrect facilitator key")
			return api.NewPermissionDeniedResponse(ctx, corsHeaders(ctx, request.Headers)), nil
		}

		sess.VotesShown = false
		for i := 0; i < len(sess.Participants); i++ {
			sess.Participants[i].CurrentVote = nil
		}

		err = saveSession(ctx, *sess)
		if err != nil {
			zerolog.Ctx(ctx).Error().Err(err).Msg("error saving session")
			return api.NewInternalServerError(ctx, corsHeaders(ctx, request.Headers)), nil
		}
		return api.NewNoContentResponse(ctx, corsHeaders(ctx, request.Headers)), nil
	}
}

func main() {
	lambdautil.CoreStartup()

	logPreparer := logging.NewPreparer()
	sess := lambdautil.DefaultAWSConfig()

	dynamo := lambdautil.NewDynamoClient(sess)
	loader := session.NewLoader(dynamo, lambdautil.SessionTable)
	notifier := session.NewChangeNotifier(dynamo, lambdautil.SessionTable, lambdautil.NewProdMessageDispatcher())
	saveSess := session.NewSaver(dynamo, lambdautil.SessionTable, notifier, lambdautil.SessionTimeout)

	allowedDomains := lambdautil.AllowedCORSOrigins()

	lambda.Start(NewHandler(logPreparer, cors.NewResponseHeaderBuilder(allowedDomains), loader, saveSess))
}
