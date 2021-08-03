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

type LoadResponse struct {
	Session    session.CompleteSessionView `json:"session"`
}

func NewHandler(prepareLogs logging.Preparer, corsHeaders cors.ResponseHeaderBuilder, loadSession session.Loader, dispatch api.MessageDispatcher, saveUser session.UserSaver) func(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	return func(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
		ctx = prepareLogs(ctx)

		l := new(session.SetFacilitatorSessionRequest)
		err := json.Unmarshal([]byte(request.Body), l)
		if err != nil {
			zerolog.Ctx(ctx).Warn().Str("error", fmt.Sprintf("%+v", err)).Msg("error reading load request body")
			return api.NewInternalServerError(ctx, corsHeaders(ctx, request.Headers)), nil
		}

		sessionID := request.PathParameters["session"]
		sess, err := loadSession(ctx, sessionID)
		if err != nil {
			zerolog.Ctx(ctx).Error().Str("error", fmt.Sprintf("%+v", err)).Msg("error reading session")
			return api.NewInternalServerError(ctx, corsHeaders(ctx, request.Headers)), nil
		}
		if sess == nil {
			zerolog.Ctx(ctx).Warn().Str("sessionID", sessionID).Msg("session not found")
		}

		facilitatorKey := api.FacilitatorKey(request.Headers)
		if sess.FacilitatorSessionKey != facilitatorKey {
			zerolog.Ctx(ctx).Warn().Msg("attempt to load session as facilitator with invalid facilitator key")
			return api.NewPermissionDeniedResponse(ctx, corsHeaders(ctx, request.Headers)), nil
		}

		sess.Facilitator.SocketID = l.ConnectionID
		err = saveUser(ctx, sessionID, sess.Facilitator, session.Facilitator)
		if err != nil {
			zerolog.Ctx(ctx).Error().Str("error", fmt.Sprintf("%+v", err)).Msg("error saving session")
			return api.NewInternalServerError(ctx, corsHeaders(ctx, request.Headers)), nil
		}

		err = dispatch(ctx, l.ConnectionID, api.Message{
			Type: api.FacilitatorSessionLoaded,
			Body: LoadResponse{
				Session:    *sess,
			},
		})
		if err != nil {
			zerolog.Ctx(ctx).Error().Str("error", fmt.Sprintf("%+v", err)).Msg("error dispatching message")
		}
		return api.NewSuccessResponse(ctx, corsHeaders(ctx, request.Headers), sess), nil
	}
}

func main() {
	lambdautil.CoreStartup()

	logPreparer := logging.NewPreparer()
	sess := lambdautil.DefaultAWSConfig()

	dynamo := lambdautil.NewDynamoClient(sess)
	loader := session.NewLoader(dynamo, lambdautil.SessionTable)
	dispatcher := lambdautil.NewProdMessageDispatcher()
	userSaver := session.NewUserSaver(dynamo, lambdautil.SessionTable, lambdautil.SessionTimeout)

	allowedDomains := strings.Split(os.Getenv("ALLOWED_ORIGINS"), ",")

	lambda.Start(NewHandler(logPreparer, cors.NewResponseHeaderBuilder(allowedDomains), loader, dispatcher, userSaver))
}
