package main

import (
	"context"
	"os"
	"strings"

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

func NewHandler(prepareLogs logging.Preparer, corsHeaders cors.ResponseHeaderBuilder, fetchProfile profile.Fetcher) func(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	return func(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
		ctx = prepareLogs(ctx)
		principal, err := aws.ExtractPrincipal(request)
		if err != nil {
			zerolog.Ctx(ctx).Err(err).Stack().Msg("error extracting principal")
			return api.NewInternalServerError(ctx, corsHeaders(ctx, request.Headers)), nil
		}

		p, err := fetchProfile(ctx, principal.UserID)
		if err != nil {
			zerolog.Ctx(ctx).Err(err).Stack().Msg("error fetching profile")
			return api.NewInternalServerError(ctx, corsHeaders(ctx, request.Headers)), nil
		}
		if p == nil {
			zerolog.Ctx(ctx).Err(err).Stack().Msg("profile not found")
			return api.NewInternalServerError(ctx, corsHeaders(ctx, request.Headers)), nil
		}
		return api.NewSuccessResponse(ctx, corsHeaders(ctx, request.Headers), profile.UserView{
			Email:  p.Email,
			Name:   p.Name,
			Handle: p.Handle,
		}), nil
	}
}

func main() {
	lambdautil.CoreStartup()

	logPreparer := logging.NewPreparer()

	sess := lambdautil.DefaultAWSConfig()

	profileTable := lambdautil.ProfileTable
	dynamo := lambdautil.NewDynamoClient(sess)
	fetchProfile := profile.NewFetcher(dynamo, profileTable)

	allowedDomains := strings.Split(os.Getenv("ALLOWED_ORIGINS"), ",")

	handler := NewHandler(logPreparer, cors.NewResponseHeaderBuilder(allowedDomains), fetchProfile)
	lambda.Start(handler)
}
