package main

import (
	"context"
	"net/http"
	"os"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/jonsabados/goauth"
	"github.com/jonsabados/goauth/aws"
	"github.com/jonsabados/goauth/google"
	"github.com/rs/zerolog"

	"github.com/jonsabados/pointypoints/logging"
)

type authCallback struct {
	ctx context.Context
}

func (a *authCallback) AuthFailed() {
	zerolog.Ctx(a.ctx).Info().Msg("authentication failed")
}

func (a *authCallback) AuthPass(p goauth.Principal) {
	zerolog.Ctx(a.ctx).Info().Str("email", p.Email).Str("id", p.UserID).Msg("auth passed")
}

type endpointMapper struct {
}

func (e *endpointMapper) AllowedEndpoints(_ context.Context, _ goauth.Principal) ([]aws.AllowedEndpoint, error) {
	return []aws.AllowedEndpoint{
		{
			Method: http.MethodGet,
			Path:   "profile",
		},
		{
			Method: http.MethodPut,
			Path:   "profile",
		},
	}, nil
}

func main() {
	logPreparer := logging.NewPreparer()

	googleClientID := os.Getenv("GOOGLE_CLIENT_ID")

	conf := aws.AuthorizerLambdaConfig{}
	conf.CallbackFactory = func(ctx context.Context) aws.AuthorizerCallback {
		ctx = logPreparer(ctx)
		return &authCallback{ctx}
	}

	certFetcher := google.NewCachingCertFetcher(google.NewCertFetcher(aws.NewXRAYAwareHTTPClientFactory(http.DefaultClient)))
	authorizer := google.NewWebSignInTokenAuthenticator(certFetcher, googleClientID)
	conf.Authorizer = authorizer

	region := os.Getenv("AWS_REGION")
	accountID := os.Getenv("ACCOUNT_ID")
	apiID := os.Getenv("API_ID")
	stage := os.Getenv("STAGE")
	conf.PolicyBuilder = aws.NewGatewayPolicyBuilder(region, accountID, apiID, stage, &endpointMapper{})

	handler := aws.NewAuthorizerLambdaHandler(conf)
	lambda.Start(handler)
}
