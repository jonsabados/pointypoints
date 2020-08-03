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
	"github.com/jonsabados/pointypoints/logging"
	"github.com/jonsabados/pointypoints/session"
	"github.com/rs/zerolog"
	"os"
	"time"
)

func NewHandler(prepareLogs logging.Preparer, startSession session.Starter) func(ctx context.Context, request events.APIGatewayWebsocketProxyRequest) api.Response {
	return func(ctx context.Context, request events.APIGatewayWebsocketProxyRequest) api.Response {
		ctx = prepareLogs(ctx)
		toStart := new(session.StartRequest)
		err := json.Unmarshal([]byte(request.Body), toStart)
		if err != nil {
			zerolog.Ctx(ctx).Warn().Str("error", fmt.Sprintf("%+v", err)).Msg("error session start reading request body")
			return api.NewInternalServerError(ctx)
		}
		if toStart.Facilitator.Name == "" {
			return api.NewValidationFailureResponse(ctx, api.ValidationError{
				Errors: []string{"facilitator name is required"},
			})
		}
		toStart.Facilitator.SocketID = request.RequestContext.ConnectionID

		sess, err := startSession(ctx, *toStart)
		if err != nil {
			zerolog.Ctx(ctx).Warn().Str("error", fmt.Sprintf("%+v", err)).Msg("error starting session")
			return api.NewInternalServerError(ctx)
		}
		return api.NewSuccessResponse(ctx, sess)
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

	lambda.Start(NewHandler(logPreparer, starter))
}
