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

type LoadResponse struct {
	Session    session.CompleteSessionView `json:"session"`
	MarkActive bool                        `json:"markActive"`
}

func NewHandler(prepareLogs logging.Preparer, loadSession session.Loader, dispatch api.MessageDispatcher, recordInterest session.InterestRecorder, locker lock.GlobalLockAppropriator, saveSession session.Saver) func(ctx context.Context, request events.APIGatewayWebsocketProxyRequest) (events.APIGatewayProxyResponse, error) {
	return func(ctx context.Context, request events.APIGatewayWebsocketProxyRequest) (events.APIGatewayProxyResponse, error) {
		ctx = prepareLogs(ctx)
		l := new(session.LoadFacilitatorSessionRequest)
		err := json.Unmarshal([]byte(request.Body), l)
		if err != nil {
			zerolog.Ctx(ctx).Warn().Str("error", fmt.Sprintf("%+v", err)).Msg("error reading load request body")
			return api.NewInternalServerError(ctx), nil
		}

		// will need to update the facilitator connection id
		lck, err := locker(ctx, lock.SessionLockKey(l.SessionID))
		if err != nil {
			zerolog.Ctx(ctx).Error().Str("error", fmt.Sprintf("%+v", err)).Msg("error locking session")
			return api.NewInternalServerError(ctx), nil
		}
		defer func() {
			err := lck.Unlock(ctx)
			if err != nil {
				zerolog.Ctx(ctx).Error().Str("errpr", fmt.Sprintf("%+v", err)).Msg("error releasing lock")
			}
		}()
		sess, err := loadSession(ctx, l.SessionID)
		if err != nil {
			zerolog.Ctx(ctx).Error().Str("error", fmt.Sprintf("%+v", err)).Msg("error reading session")
			return api.NewInternalServerError(ctx), nil
		}
		if sess == nil {
			zerolog.Ctx(ctx).Warn().Str("sessionID", l.SessionID).Msg("session not found")
		}
		if sess.FacilitatorSessionKey != l.FacilitatorSessionKey {
			zerolog.Ctx(ctx).Warn().Msg("attempt to load session as facilitator with invalid facilitator key")
			return api.NewPermissionDeniedResponse(ctx), nil
		}
		sess.Facilitator.SocketID = request.RequestContext.ConnectionID
		err = saveSession(ctx, *sess)
		if err != nil {
			zerolog.Ctx(ctx).Error().Str("error", fmt.Sprintf("%+v", err)).Msg("error saving session")
			return api.NewInternalServerError(ctx), nil
		}


		err = recordInterest(ctx, sess.SessionID, request.RequestContext.ConnectionID)
		if err != nil {
			zerolog.Ctx(ctx).Error().Str("error", fmt.Sprintf("%+v", err)).Msg("error recording interest")
			return api.NewInternalServerError(ctx), nil
		}
		err = dispatch(ctx, request.RequestContext.ConnectionID, api.Message{
			Type: api.FacilitatorSessionLoaded,
			Body: LoadResponse{
				Session:    *sess,
				MarkActive: l.MarkActive,
			},
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
	loader := session.NewLoader(dynamo, sessionTable)

	lockTable := os.Getenv("LOCK_TABLE")
	locker := lock.NewGlobalLockAppropriator(dynamo, lockTable, time.Millisecond*5, time.Second)

	interestTable := os.Getenv("INTEREST_TABLE")
	watcherTable := os.Getenv("WATCHER_TABLE")
	interestRecorder := session.NewInterestRecorder(dynamo, interestTable, watcherTable, locker, time.Hour)

	dispatcher := diutil.NewProdMessageDispatcher()

	notifier := session.NewChangeNotifier(dynamo, watcherTable, dispatcher)
	saveSess := session.NewSaver(dynamo, sessionTable, notifier, time.Hour)

	lambda.Start(NewHandler(logPreparer, loader, dispatcher, interestRecorder, locker, saveSess))
}
