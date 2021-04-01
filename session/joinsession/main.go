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

func NewHandler(prepareLogs logging.Preparer, loadSession session.Loader, saveUser session.UserSaver, notifyParticipants session.ChangeNotifier) func(ctx context.Context, request events.APIGatewayWebsocketProxyRequest) (events.APIGatewayProxyResponse, error) {
	return func(ctx context.Context, request events.APIGatewayWebsocketProxyRequest) (events.APIGatewayProxyResponse, error) {
		ctx = prepareLogs(ctx)
		j := new(session.JoinSessionRequest)
		err := json.Unmarshal([]byte(request.Body), j)
		if err != nil {
			zerolog.Ctx(ctx).Warn().Str("error", fmt.Sprintf("%+v", err)).Msg("error reading load request body")
			return api.NewInternalServerError(ctx), nil
		}

		errors := make([]string, 0)
		if j.User.Name == "" {
			errors = append(errors, "user name is required")
		}
		if j.User.UserID == "" {
			errors = append(errors, "user id is required")
		}
		if len(errors) > 0 {
			return api.NewValidationFailureResponse(ctx, api.ValidationError{
				Errors: errors,
			}), nil
		}

		newUser := j.User
		newUser.SocketID = request.RequestContext.ConnectionID
		err = saveUser(ctx, j.SessionID, newUser, session.Participant)
		if err != nil {
			zerolog.Ctx(ctx).Error().Str("error", fmt.Sprintf("%+v", err)).Msg("error saving session")
			return api.NewInternalServerError(ctx), nil
		}

		sess, err := loadSession(ctx, j.SessionID)
		if err != nil {
			zerolog.Ctx(ctx).Error().Str("error", fmt.Sprintf("%+v", err)).Msg("error reading session")
			return api.NewPermissionDeniedResponse(ctx), nil
		}
		err = notifyParticipants(ctx, *sess)
		if err != nil {
			zerolog.Ctx(ctx).Error().Str("error", fmt.Sprintf("%+v", err)).Msg("error notifying participants of change")
			return api.NewPermissionDeniedResponse(ctx), nil
		}

		return api.NewSuccessResponse(ctx, sess), nil
	}
}

func main() {
	lambdautil.CoreStartup()

	logPreparer := logging.NewPreparer()
	sess := lambdautil.DefaultAWSConfig()

	dynamo := lambdautil.NewDynamoClient(sess)
	loader := session.NewLoader(dynamo, lambdautil.SessionTable)
	saver := session.NewUserSaver(dynamo, lambdautil.SessionTable, lambdautil.SessionTimeout)
	notifier := session.NewChangeNotifier(dynamo, lambdautil.SessionTable, lambdautil.NewProdMessageDispatcher())

	lambda.Start(NewHandler(logPreparer, loader, saver, notifier))
}
