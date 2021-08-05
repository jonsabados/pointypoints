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

func NewHandler(prepareLogs logging.Preparer, corsHeaders cors.ResponseHeaderBuilder, loadSession session.Loader, saveUser session.UserSaver, notifyParticipants session.ChangeNotifier) func(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	return func(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
		ctx = prepareLogs(ctx)
		var joinRequest session.JoinSessionRequest
		err := json.Unmarshal([]byte(request.Body), &joinRequest)
		if err != nil {
			zerolog.Ctx(ctx).Warn().Err(err).Msg("error reading load request body")
			return api.NewInternalServerError(ctx, corsHeaders(ctx, request.Headers)), nil
		}

		errors := make([]string, 0)
		if joinRequest.Name == "" {
			errors = append(errors, "user name is required")
		}
		if joinRequest.ConnectionID == "" {
			errors = append(errors, "connection id is required")
		}
		if len(errors) > 0 {
			return api.NewValidationFailureResponse(ctx, corsHeaders(ctx, request.Headers), api.ValidationError{
				Errors: errors,
			}), nil
		}

		sessionID := request.PathParameters["session"]
		user := session.User{
			UserID:   request.PathParameters["user"],
			Name:     joinRequest.Name,
			Handle:   joinRequest.Handle,
			SocketID: joinRequest.ConnectionID,
		}

		principal, err := aws.ExtractPrincipal(request)
		if err != nil {
			zerolog.Ctx(ctx).Warn().Err(err).Msg("error extracting principal")
			return api.NewInternalServerError(ctx, corsHeaders(ctx, request.Headers)), nil
		}

		err = saveUser(ctx, principal, sessionID, user, session.Participant, false)
		if err != nil {
			zerolog.Ctx(ctx).Error().Err(err).Msg("error saving session")
			return api.NewInternalServerError(ctx, corsHeaders(ctx, request.Headers)), nil
		}

		sess, err := loadSession(ctx, sessionID)
		if err != nil {
			zerolog.Ctx(ctx).Error().Err(err).Msg("error reading session")
			return api.NewPermissionDeniedResponse(ctx, corsHeaders(ctx, request.Headers)), nil
		}
		err = notifyParticipants(ctx, *sess)
		if err != nil {
			zerolog.Ctx(ctx).Error().Err(err).Msg("error notifying participants of change")
			return api.NewPermissionDeniedResponse(ctx, corsHeaders(ctx, request.Headers)), nil
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
	saver := session.NewUserSaver(dynamo, lambdautil.SessionTable, lambdautil.SessionTimeout, statsFactory)
	notifier := session.NewChangeNotifier(dynamo, lambdautil.SessionTable, lambdautil.NewProdMessageDispatcher())

	allowedDomains := strings.Split(os.Getenv("ALLOWED_ORIGINS"), ",")

	lambda.Start(NewHandler(logPreparer, cors.NewResponseHeaderBuilder(allowedDomains), loader, saver, notifier))
}
