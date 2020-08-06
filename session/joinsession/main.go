package main

import (
	"context"
	"encoding/json"
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

func NewHandler(prepareLogs logging.Preparer, loadSession session.Loader, locker lock.GlobalLockAppropriator, saveSession session.Saver) func(ctx context.Context, request events.APIGatewayWebsocketProxyRequest) (events.APIGatewayProxyResponse, error) {
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

		recordLock, err := locker(ctx, lock.SessionLockKey(j.SessionID))
		if err != nil {
			zerolog.Ctx(ctx).Error().Str("error", fmt.Sprintf("%+v", err)).Msg("error locking session")
			return api.NewInternalServerError(ctx), nil
		}
		defer func() {
			err := recordLock.Unlock(ctx)
			if err != nil {
				zerolog.Ctx(ctx).Error().Str("error", fmt.Sprintf("%+v", err)).Msg("unable to release lock")
			}
		}()
		sess, err := loadSession(ctx, j.SessionID)
		if err != nil {
			zerolog.Ctx(ctx).Error().Str("error", fmt.Sprintf("%+v", err)).Msg("error reading session")
			return api.NewPermissionDeniedResponse(ctx), nil
		}
		if sess == nil {
			zerolog.Ctx(ctx).Warn().Str("sessionID", j.SessionID).Msg("session not found")
			return api.NewPermissionDeniedResponse(ctx), nil
		}

		newUser := j.User
		newUser.SocketID = request.RequestContext.ConnectionID
		sess.Participants = append(sess.Participants, newUser)
		err = saveSession(ctx, *sess)
		if err != nil {
			zerolog.Ctx(ctx).Error().Str("error", fmt.Sprintf("%+v", err)).Msg("error saving session")
			return api.NewInternalServerError(ctx), nil
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
	locker := lock.NewGlobalLockAppropriator(dynamo, lambdautil.LockTable, lambdautil.LockWaitTime, lambdautil.LockTimeout)
	notifier := session.NewChangeNotifier(dynamo, lambdautil.WatcherTable, lambdautil.NewProdMessageDispatcher())
	saveSess := session.NewSaver(dynamo, lambdautil.SessionTable, notifier, lambdautil.SessionTimeout)

	lambda.Start(NewHandler(logPreparer, loader, locker, saveSess))
}
