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

func NewHandler(prepareLogs logging.Preparer, corsHeaders cors.ResponseHeaderBuilder, startSession session.Starter) func(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	return func(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
		ctx = prepareLogs(ctx)
		toStart := new(session.StartRequest)
		err := json.Unmarshal([]byte(request.Body), toStart)
		if err != nil {
			zerolog.Ctx(ctx).Warn().Str("error", fmt.Sprintf("%+v", err)).Msg("error session start reading request body")
			return api.NewInternalServerError(ctx, corsHeaders(request.Headers)), nil
		}
		errors := make([]string, 0)
		if toStart.Facilitator.Name == "" {
			errors = append(errors, "facilitator name is required")
		}
		if toStart.Facilitator.UserID == "" {
			errors = append(errors, "facilitator user id is required")
		}
		if toStart.ConnectionID == "" {
			errors = append(errors, "connection id is required")
		}
		if len(errors) > 0 {
			return api.NewValidationFailureResponse(ctx, corsHeaders(request.Headers), api.ValidationError{
				Errors: errors,
			}), nil
		}

		toStart.Facilitator.SocketID = toStart.ConnectionID

		sess, err := startSession(ctx, *toStart)
		if err != nil {
			zerolog.Ctx(ctx).Warn().Str("error", fmt.Sprintf("%+v", err)).Msg("error starting session")
			return api.NewInternalServerError(ctx, corsHeaders(request.Headers)), nil
		}

		return api.NewSuccessResponse(ctx, corsHeaders(request.Headers), sess), nil
	}
}

func main() {
	lambdautil.CoreStartup()

	logPreparer := logging.NewPreparer()
	sess := lambdautil.DefaultAWSConfig()

	dynamo := lambdautil.NewDynamoClient(sess)
	starter := session.NewStarter(dynamo, lambdautil.SessionTable, lambdautil.SessionTimeout)

	allowedDomains := strings.Split(os.Getenv("ALLOWED_ORIGINS"), ",")

	lambda.Start(NewHandler(logPreparer, cors.NewResponseHeaderBuilder(allowedDomains), starter))
}
