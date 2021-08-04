package main

import (
	"context"
	"net/http"
	"os"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/jonsabados/goauth"
	"github.com/jonsabados/goauth/aws"
	"github.com/jonsabados/goauth/google"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"

	"github.com/jonsabados/pointypoints/lambdautil"
	"github.com/jonsabados/pointypoints/logging"
	"github.com/jonsabados/pointypoints/profile"
)

type authCallback struct {
	ctx          context.Context
	fetchProfile profile.Fetcher
	writeProfile profile.Writer
}

func (a *authCallback) ErrorEncountered(err error) {
	zerolog.Ctx(a.ctx).Error().Err(err).Stack().Msg("error encountered")
}

func (a *authCallback) AuthFailed() error {
	zerolog.Ctx(a.ctx).Info().Msg("authentication failed")
	return nil
}

func (a *authCallback) AuthPass(p goauth.Principal) error {
	zerolog.Ctx(a.ctx).Info().Str("email", p.Email).Str("id", p.UserID).Msg("auth passed")
	saved, err := a.fetchProfile(a.ctx, p.UserID)
	if err != nil {
		return err
	}
	if saved == nil {
		zerolog.Ctx(a.ctx).Info().Msg("first time user, saving profile")
		err := a.writeProfile(a.ctx, profile.Profile{
			UserID: p.UserID,
			Email:  p.Email,
			Name:   p.Name,
		})
		if err != nil {
			return errors.Wrap(err, "error writing profile")
		}
	} else if saved.Name != p.Name || saved.Email != p.Email {
		// unsure if you can update name or email with google accounts... but lets support it just in case.
		err := a.writeProfile(a.ctx, profile.Profile{
			UserID: p.UserID,
			Email:  p.Email,
			Name:   p.Name,
			Handle: saved.Handle,
		})
		if err != nil {
			return errors.Wrap(err, "error writing profile")
		}
	}
	return nil
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
	lambdautil.CoreStartup()
	logPreparer := logging.NewPreparer()

	googleClientID := os.Getenv("GOOGLE_CLIENT_ID")

	sess := lambdautil.DefaultAWSConfig()

	profileTable := os.Getenv("PROFILE_TABLE")
	dynamo := lambdautil.NewDynamoClient(sess)
	fetchProfile := profile.NewFetcher(dynamo, profileTable)
	writeProfile := profile.NewWriter(dynamo, profileTable)

	conf := aws.AuthorizerLambdaConfig{}
	conf.CallbackFactory = func(ctx context.Context) aws.AuthorizerCallback {
		ctx = logPreparer(ctx)
		return &authCallback{ctx, fetchProfile, writeProfile}
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
