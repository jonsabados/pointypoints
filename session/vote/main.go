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
		r := new(session.VoteRequest)
		err := json.Unmarshal([]byte(request.Body), r)
		if err != nil {
			zerolog.Ctx(ctx).Warn().Str("error", fmt.Sprintf("%+v", err)).Msg("error reading load request body")
			return api.NewInternalServerError(ctx), nil
		}

		sess, err := loadSession(ctx, r.SessionID)
		if err != nil {
			zerolog.Ctx(ctx).Error().Str("error", fmt.Sprintf("%+v", err)).Msg("error reading session")
			return api.NewPermissionDeniedResponse(ctx), nil
		}
		if sess == nil {
			zerolog.Ctx(ctx).Warn().Str("sessionID", r.SessionID).Msg("session not found")
			return api.NewPermissionDeniedResponse(ctx), nil
		}
		zerolog.Ctx(ctx).Debug().Interface("session", sess).Msg("loaded session")

		var user *session.User
		userType := session.Participant
		if sess.FacilitatorPoints && sess.Facilitator.SocketID == request.RequestContext.ConnectionID {
			userType = session.Facilitator
			user = &sess.Facilitator
		} else {
			for i := 0; i < len(sess.Participants); i++ {
				if sess.Participants[i].SocketID == request.RequestContext.ConnectionID {
					user = &sess.Participants[i]
					break
				}
			}
		}
		if user == nil {
			return api.NewPermissionDeniedResponse(ctx), nil
		}

		user.CurrentVote = &r.Vote

		err = saveUser(ctx, r.SessionID, *user, userType)
		if err != nil {
			zerolog.Ctx(ctx).Error().Str("error", fmt.Sprintf("%+v", err)).Msg("error saving session")
			return api.NewInternalServerError(ctx), nil
		}
		err = notifyParticipants(ctx, *sess)
		if err != nil {
			zerolog.Ctx(ctx).Error().Str("error", fmt.Sprintf("%+v", err)).Msg("error notifying participants")
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
	saveUser := session.NewUserSaver(dynamo, lambdautil.SessionTable, lambdautil.SessionTimeout)
	notifier := session.NewChangeNotifier(dynamo, lambdautil.SessionTable, lambdautil.NewProdMessageDispatcher())

	lambda.Start(NewHandler(logPreparer, loader, saveUser, notifier))
}
