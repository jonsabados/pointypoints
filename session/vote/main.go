package main

import (
	"context"
	"encoding/json"
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
	"github.com/jonsabados/pointypoints/session"
)

func NewHandler(prepareLogs logging.Preparer, corsHeaders cors.ResponseHeaderBuilder, loadSession session.Loader, saveUser session.UserSaver, notifyParticipants session.ChangeNotifier) func(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	return func(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
		ctx = prepareLogs(ctx)
		r := new(session.VoteRequest)
		err := json.Unmarshal([]byte(request.Body), r)
		if err != nil {
			zerolog.Ctx(ctx).Warn().Err(err).Msg("error reading load request body")
			return api.NewInternalServerError(ctx, corsHeaders(ctx, request.Headers)), nil
		}

		// if requests made it to the lambda without a session or connection path param things have gone wrong and a panic is OK
		sessionID := request.PathParameters["session"]
		sess, err := loadSession(ctx, sessionID)
		if err != nil {
			zerolog.Ctx(ctx).Error().Err(err).Msg("error reading session")
			return api.NewPermissionDeniedResponse(ctx, corsHeaders(ctx, request.Headers)), nil
		}
		if sess == nil {
			zerolog.Ctx(ctx).Warn().Str("sessionID", sessionID).Msg("session not found")
			return api.NewPermissionDeniedResponse(ctx, corsHeaders(ctx, request.Headers)), nil
		}
		zerolog.Ctx(ctx).Debug().Interface("session", sess).Msg("loaded session")

		userID := request.PathParameters["user"]
		var user *session.User
		userType := session.Participant
		if sess.FacilitatorPoints && sess.Facilitator.UserID == userID {
			userType = session.Facilitator
			user = &sess.Facilitator
		} else {
			for i := 0; i < len(sess.Participants); i++ {
				if sess.Participants[i].UserID == userID {
					user = &sess.Participants[i]
					break
				}
			}
		}
		if user == nil {
			return api.NewPermissionDeniedResponse(ctx, corsHeaders(ctx, request.Headers)), nil
		}

		user.CurrentVote = &r.Vote

		principal, err := aws.ExtractPrincipal(request)
		if err != nil {
			zerolog.Ctx(ctx).Warn().Err(err).Msg("error extracting principal")
			return api.NewInternalServerError(ctx, corsHeaders(ctx, request.Headers)), nil
		}

		err = saveUser(ctx, principal, sessionID, *user, userType, true)
		if err != nil {
			zerolog.Ctx(ctx).Error().Err(err).Msg("error saving session")
			return api.NewInternalServerError(ctx, corsHeaders(ctx, request.Headers)), nil
		}
		err = notifyParticipants(ctx, *sess)
		if err != nil {
			zerolog.Ctx(ctx).Error().Err(err).Msg("error notifying participants")
			return api.NewInternalServerError(ctx, corsHeaders(ctx, request.Headers)), nil
		}

		return api.NewNoContentResponse(ctx, corsHeaders(ctx, request.Headers)), nil
	}
}

func main() {
	lambdautil.CoreStartup()

	logPreparer := logging.NewPreparer()
	sess := lambdautil.DefaultAWSConfig()

	statsFactory := profile.NewStatsUpdateFactory(lambdautil.ProfileTable)

	dynamo := lambdautil.NewDynamoClient(sess)
	loader := session.NewLoader(dynamo, lambdautil.SessionTable)
	saveUser := session.NewUserSaver(dynamo, lambdautil.SessionTable, lambdautil.SessionTimeout, statsFactory)
	notifier := session.NewChangeNotifier(dynamo, lambdautil.SessionTable, lambdautil.NewProdMessageDispatcher())

	allowedDomains := strings.Split(os.Getenv("ALLOWED_ORIGINS"), ",")

	lambda.Start(NewHandler(logPreparer, cors.NewResponseHeaderBuilder(allowedDomains), loader, saveUser, notifier))
}
