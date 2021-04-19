package main

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/rs/zerolog"

	"github.com/jonsabados/pointypoints/api"
	"github.com/jonsabados/pointypoints/lambdautil"
	"github.com/jonsabados/pointypoints/logging"
	"github.com/jonsabados/pointypoints/session"
)

func NewHandler(prepareLogs logging.Preparer, startSession session.Starter, dispatch api.MessageDispatcher) func(ctx context.Context, request events.APIGatewayWebsocketProxyRequest) (events.APIGatewayProxyResponse, error) {
	return func(ctx context.Context, request events.APIGatewayWebsocketProxyRequest) (events.APIGatewayProxyResponse, error) {
		ctx = prepareLogs(ctx)
		toStart := new(session.StartRequest)
		err := json.Unmarshal([]byte(request.Body), toStart)
		if err != nil {
			zerolog.Ctx(ctx).Warn().Str("error", fmt.Sprintf("%+v", err)).Msg("error session start reading request body")
			return api.NewInternalServerError(ctx, nil), nil
		}
		errors := make([]string, 0)
		if toStart.Facilitator.Name == "" {
			errors = append(errors, "facilitator name is required")

		}
		if toStart.Facilitator.UserID == "" {
			errors = append(errors, "facilitator user id is required")
		}
		if len(errors) > 0 {
			return api.NewValidationFailureResponse(ctx, nil, api.ValidationError{
				Errors: errors,
			}), nil
		}

		toStart.Facilitator.SocketID = request.RequestContext.ConnectionID

		sess, err := startSession(ctx, *toStart)
		if err != nil {
			zerolog.Ctx(ctx).Warn().Str("error", fmt.Sprintf("%+v", err)).Msg("error starting session")
			err = dispatch(ctx, request.RequestContext.ConnectionID, api.Message{
				Type: api.ErrorEncountered,
				Body: err.Error(),
			})
			if err != nil {
				zerolog.Ctx(ctx).Error().Str("error", fmt.Sprintf("%+v", err)).Msg("error dispatching message")
			}
			return api.NewInternalServerError(ctx, nil), nil
		}

		err = dispatch(ctx, request.RequestContext.ConnectionID, api.Message{
			Type: api.SessionCreated,
			Body: sess,
		})
		if err != nil {
			zerolog.Ctx(ctx).Error().Str("error", fmt.Sprintf("%+v", err)).Msg("error dispatching message")
		}
		return api.NewSuccessResponse(ctx, nil, sess), nil
	}
}

func main() {
	lambdautil.CoreStartup()

	logPreparer := logging.NewPreparer()
	sess := lambdautil.DefaultAWSConfig()

	dynamo := lambdautil.NewDynamoClient(sess)
	starter := session.NewStarter(dynamo, lambdautil.SessionTable, time.Hour)

	lambda.Start(NewHandler(logPreparer, starter, lambdautil.NewProdMessageDispatcher()))
}
