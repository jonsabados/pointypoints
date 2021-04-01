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

type LoadResponse struct {
	Session    session.CompleteSessionView `json:"session"`
	MarkActive bool                        `json:"markActive"`
}

func NewHandler(prepareLogs logging.Preparer, loadSession session.Loader, dispatch api.MessageDispatcher, saveUser session.UserSaver) func(ctx context.Context, request events.APIGatewayWebsocketProxyRequest) (events.APIGatewayProxyResponse, error) {
	return func(ctx context.Context, request events.APIGatewayWebsocketProxyRequest) (events.APIGatewayProxyResponse, error) {
		ctx = prepareLogs(ctx)
		l := new(session.LoadFacilitatorSessionRequest)
		err := json.Unmarshal([]byte(request.Body), l)
		if err != nil {
			zerolog.Ctx(ctx).Warn().Str("error", fmt.Sprintf("%+v", err)).Msg("error reading load request body")
			return api.NewInternalServerError(ctx), nil
		}

		sess, err := loadSession(ctx, l.SessionID)
		if err != nil {
			zerolog.Ctx(ctx).Error().Str("error", fmt.Sprintf("%+v", err)).Msg("error reading session")
			return api.NewInternalServerError(ctx), nil
		}
		if sess == nil {
			zerolog.Ctx(ctx).Warn().Str("sessionID", l.SessionID).Msg("session not found")
		}
		if sess.FacilitatorSessionKey != l.FacilitatorSessionKey {
			zerolog.Ctx(ctx).Warn().Msg("attempt to load session as facilitator with invalid facilitator key")
			return api.NewPermissionDeniedResponse(ctx), nil
		}

		sess.Facilitator.SocketID = request.RequestContext.ConnectionID
		err = saveUser(ctx, l.SessionID, sess.Facilitator, session.Facilitator)
		if err != nil {
			zerolog.Ctx(ctx).Error().Str("error", fmt.Sprintf("%+v", err)).Msg("error saving session")
			return api.NewInternalServerError(ctx), nil
		}

		err = dispatch(ctx, request.RequestContext.ConnectionID, api.Message{
			Type: api.FacilitatorSessionLoaded,
			Body: LoadResponse{
				Session:    *sess,
				MarkActive: l.MarkActive,
			},
		})
		if err != nil {
			zerolog.Ctx(ctx).Error().Str("error", fmt.Sprintf("%+v", err)).Msg("error dispatching message")
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
	dispatcher := lambdautil.NewProdMessageDispatcher()
	userSaver := session.NewUserSaver(dynamo, lambdautil.SessionTable, lambdautil.SessionTimeout)

	lambda.Start(NewHandler(logPreparer, loader, dispatcher, userSaver))
}
