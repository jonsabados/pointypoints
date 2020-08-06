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
	"github.com/jonsabados/pointypoints/lock"
	"github.com/jonsabados/pointypoints/logging"
	"github.com/jonsabados/pointypoints/session"
	"github.com/rs/zerolog"
	"os"
	"time"
)

func NewHandler(prepareLogs logging.Preparer, startSession session.Starter, dispatch api.MessageDispatcher, recordInterest session.InterestRecorder) func(ctx context.Context, request events.APIGatewayWebsocketProxyRequest) (events.APIGatewayProxyResponse, error) {
	return func(ctx context.Context, request events.APIGatewayWebsocketProxyRequest) (events.APIGatewayProxyResponse, error) {
		ctx = prepareLogs(ctx)
		toStart := new(session.StartRequest)
		err := json.Unmarshal([]byte(request.Body), toStart)
		if err != nil {
			zerolog.Ctx(ctx).Warn().Str("error", fmt.Sprintf("%+v", err)).Msg("error session start reading request body")
			return api.NewInternalServerError(ctx), nil
		}
		errors := make([]string, 0)
		if toStart.Facilitator.Name == "" {
			errors = append(errors, "facilitator name is required")

		}
		if toStart.Facilitator.UserID == "" {
			errors = append(errors, "facilitator user id is required")
		}
		if len(errors) > 0 {
			return api.NewValidationFailureResponse(ctx, api.ValidationError{
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
			return api.NewInternalServerError(ctx), nil
		}
		err = recordInterest(ctx, sess.SessionID, request.RequestContext.ConnectionID)
		if err != nil {
			zerolog.Ctx(ctx).Error().Str("error", fmt.Sprintf("%+v", err)).Msg("error recording interest")
			return api.NewInternalServerError(ctx), nil
		}
		err = dispatch(ctx, request.RequestContext.ConnectionID, api.Message{
			Type: api.SessionCreated,
			Body: sess,
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
	starter := session.NewStarter(dynamo, sessionTable, time.Hour)

	lockTable := os.Getenv("LOCK_TABLE")
	locker := lock.NewGlobalLockAppropriator(dynamo, lockTable, time.Millisecond*5, time.Second)

	interestTable := os.Getenv("INTEREST_TABLE")
	watcherTable := os.Getenv("WATCHER_TABLE")
	interestRecorder := session.NewInterestRecorder(dynamo, interestTable, watcherTable, locker, time.Hour)

	lambda.Start(NewHandler(logPreparer, starter, diutil.NewProdMessageDispatcher(), interestRecorder))
}
