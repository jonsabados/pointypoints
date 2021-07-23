package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

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
		r := new(session.UpdateRequest)
		err := json.Unmarshal([]byte(request.Body), r)
		if err != nil {
			zerolog.Ctx(ctx).Warn().Str("error", fmt.Sprintf("%+v", err)).Msg("error reading load request body")
			return api.NewInternalServerError(ctx, corsHeaders(request.Headers)), nil
		}

		sessionID := request.PathParameters["session"]

		sess, err := loadSession(ctx, sessionID)
		if err != nil {
			zerolog.Ctx(ctx).Error().Str("error", fmt.Sprintf("%+v", err)).Msg("error reading session")
			return api.NewPermissionDeniedResponse(ctx, corsHeaders(request.Headers)), nil
		}
		if sess == nil {
			zerolog.Ctx(ctx).Warn().Str("sessionID", sessionID).Msg("session not found")
			return api.NewPermissionDeniedResponse(ctx, corsHeaders(request.Headers)), nil
		}
		if facilitatorKey := api.FacilitatorKey(request.Headers); sess.FacilitatorSessionKey != facilitatorKey {
			zerolog.Ctx(ctx).Warn().Str("sessionID", sessionID).Msg("attempt to show votes with incorrect facilitator key")
			return api.NewPermissionDeniedResponse(ctx, corsHeaders(request.Headers)), nil
		}

		sess.VotesShown = r.VotesShown
		sess.FacilitatorPoints = r.FacilitatorPoints

		err = saveSession(ctx, *sess)
		if err != nil {
			zerolog.Ctx(ctx).Error().Str("error", fmt.Sprintf("%+v", err)).Msg("error saving session")
			return api.NewInternalServerError(ctx, corsHeaders(request.Headers)), nil
		}
		return api.NewSuccessResponse(ctx, corsHeaders(request.Headers), "session updated"), nil
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

	allowedDomains := strings.Split(os.Getenv("ALLOWED_ORIGINS"), ",")

	lambda.Start(NewHandler(logPreparer, cors.NewResponseHeaderBuilder(allowedDomains), loader, saveSess))
}
