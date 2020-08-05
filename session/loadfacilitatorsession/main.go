package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	awssession "github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-xray-sdk-go/xray"
	"github.com/jonsabados/pointypoints/api"
	"github.com/jonsabados/pointypoints/diutil"
	"github.com/jonsabados/pointypoints/logging"
	"github.com/jonsabados/pointypoints/session"
	"github.com/rs/zerolog"
	"os"
)

type LoadResponse struct {
	Session    session.FacilitatorSessionVew `json:"session"`
	MarkActive bool                          `json:"markActive"`
}

func NewHandler(prepareLogs logging.Preparer, loadSession session.Loader, dispatch api.MessageDispatcher) func(ctx context.Context, request events.APIGatewayWebsocketProxyRequest) (events.APIGatewayProxyResponse, error) {
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
			return api.NewPermissionDeniedResponse(ctx), nil
		}
		if sess == nil {
			zerolog.Ctx(ctx).Warn().Str("sessionID", l.SessionID).Msg("session not found")
		}
		if sess.FacilitatorSessionKey != l.FacilitatorSessionKey {
			zerolog.Ctx(ctx).Warn().Msg("attempt to load session as facilitator with invalid facilitator key")
			return api.NewPermissionDeniedResponse(ctx), nil
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
	err := xray.Configure(xray.Config{
		LogLevel: "warn",
	})
	if err != nil {
		panic(err)
	}

	logPreparer := logging.NewPreparer()
	sess, err := awssession.NewSession(&aws.Config{})
	if err != nil {
		panic(err)
	}
	dynamo := dynamodb.New(sess)
	xray.AWS(dynamo.Client)

	sessionTable := os.Getenv("SESSION_TABLE")
	loader := session.NewLoader(dynamo, sessionTable)

	lambda.Start(NewHandler(logPreparer, loader, diutil.NewProdMessageDispatcher()))
}
