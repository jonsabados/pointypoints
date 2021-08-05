package logging

import (
	"context"
	"os"

	"github.com/aws/aws-lambda-go/lambdacontext"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/pkgerrors"
)

func init() {
	zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack
}

// Preparer sets up a context for logging, returning a context that has a logger established as well as the set logger
type Preparer func(ctx context.Context) context.Context

func NewPreparer() Preparer {
	logLevel, err := zerolog.ParseLevel(os.Getenv("LOG_LEVEL"))
	if err != nil {
		tmpLogger := zerolog.New(os.Stdout)
		tmpLogger.Fatal().Err(err).Msg("unable to configure logger, set LOG_LEVEL to an appropriate value")
	}
	baseLogger := zerolog.New(os.Stdout).Level(logLevel).With().Stack().Logger()

	return func(ctx context.Context) context.Context {
		if awsCtx, inLambda := lambdacontext.FromContext(ctx); inLambda {
			logger := baseLogger.With().Str("requestId", awsCtx.AwsRequestID).Logger()
			return logger.WithContext(ctx)
		} else {
			return ctx
		}
	}
}