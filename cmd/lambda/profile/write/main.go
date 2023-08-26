package main

import (
	"context"
	"encoding/json"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/jonsabados/goauth/aws"
	"github.com/rs/zerolog"

	"github.com/jonsabados/pointypoints/api"
	"github.com/jonsabados/pointypoints/cors"
	"github.com/jonsabados/pointypoints/lambdautil"
	"github.com/jonsabados/pointypoints/logging"
	"github.com/jonsabados/pointypoints/profile"
)

func NewHandler(prepareLogs logging.Preparer, corsHeaders cors.ResponseHeaderBuilder, writeProfile profile.Writer) func(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	return func(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
		ctx = prepareLogs(ctx)
		principal, err := aws.ExtractPrincipal(request)
		if err != nil {
			zerolog.Ctx(ctx).Err(err).Stack().Msg("error extracting principal")
			return api.NewInternalServerError(ctx, corsHeaders(ctx, request.Headers)), nil
		}

		input := profile.UserView{}
		err = json.Unmarshal([]byte(request.Body), &input)
		if err != nil {
			zerolog.Ctx(ctx).Warn().Err(err).Msg("error reading request body")
			return api.NewInternalServerError(ctx, corsHeaders(ctx, request.Headers)), nil
		}

		if input.Handle != nil && *input.Handle == "" {
			input.Handle = nil
		}

		err = writeProfile(ctx, profile.Profile{
			UserID: principal.UserID,
			Email:  principal.Email,
			Name:   principal.Name,
			Handle: input.Handle,
		})
		if err != nil {
			zerolog.Ctx(ctx).Error().Err(err).Msg("error writing profile")
			return api.NewInternalServerError(ctx, corsHeaders(ctx, request.Headers)), nil
		}

		return api.NewNoContentResponse(ctx, corsHeaders(ctx, request.Headers)), nil
	}
}

func main() {
	lambdautil.CoreStartup()

	logPreparer := logging.NewPreparer()

	sess := lambdautil.DefaultAWSConfig()

	dynamo := lambdautil.NewDynamoClient(sess)
	writeProfile := profile.NewWriter(dynamo, lambdautil.ProfileTable)

	allowedDomains := lambdautil.AllowedCORSOrigins()

	handler := NewHandler(logPreparer, cors.NewResponseHeaderBuilder(allowedDomains), writeProfile)
	lambda.Start(handler)
}
