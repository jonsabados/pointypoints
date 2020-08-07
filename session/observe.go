package session

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/jonsabados/pointypoints/api"
	"github.com/jonsabados/pointypoints/lock"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"strconv"
	"time"
)

type ChangeNotifier func(ctx context.Context, updated CompleteSessionView) error

func NewChangeNotifier(dynamo DynamoClient, watcherTable string, dispatchMessage api.MessageDispatcher) ChangeNotifier {
	return func(ctx context.Context, updated CompleteSessionView) error {
		watcherRes, err := dynamo.GetItemWithContext(ctx, &dynamodb.GetItemInput{
			TableName: aws.String(watcherTable),
			Key: map[string]*dynamodb.AttributeValue{
				"SessionID": {S: aws.String(updated.SessionID)},
			},
		})
		if err != nil {
			return errors.WithStack(err)
		}
		if watcherRes.Item["SessionID"] == nil {
			return errors.New("session not found")
		}
		for _, r := range watcherRes.Item["Connections"].L {
			err := dispatchMessage(ctx, *r.S, api.Message{
				Type: api.SessionUpdated,
				Body: connectionView(updated, *r.S),
			})
			if err != nil {
				zerolog.Ctx(ctx).Warn().Str("error", fmt.Sprintf("%+v", err)).Msg("error notifying observer")
			}
		}
		return nil
	}
}

func connectionView(sess CompleteSessionView, connectionID string) interface{} {
	if sess.Facilitator.SocketID == connectionID {
		return sess
	}
	return ToParticipantView(sess, connectionID)
}

type Disconnector func(ctx context.Context, connectionID string) error

func NewDisconnector(dynamo DynamoClient, tableName string, locker lock.GlobalLockAppropriator, loadSession Loader, saveSession Saver) Disconnector {
	return func(ctx context.Context, connectionID string) error {
		rec, err := dynamo.GetItemWithContext(ctx, &dynamodb.GetItemInput{
			TableName: aws.String(tableName),
			Key: map[string]*dynamodb.AttributeValue{
				"ConnectionID": {S: aws.String(connectionID)},
			},
		})
		if err != nil {
			return errors.WithStack(err)
		}
		if rec.Item["ConnectionID"] == nil {
			zerolog.Ctx(ctx).Debug().Str("connectionID", connectionID).Msg("no records found for connection")
			return nil
		}
		for _, r := range rec.Item["Sessions"].L {
			sessionID := *r.S
			err := locker.DoWithLock(ctx, lock.SessionLockKey(sessionID), func(ctx context.Context) error {
				sess, err := loadSession(ctx, sessionID)
				if err != nil {
					return errors.WithStack(err)
				}
				newUsers := make([]User, 0)
				for _, u := range sess.Participants {
					if u.SocketID != connectionID {
						newUsers = append(newUsers, u)
					} else {
						zerolog.Ctx(ctx).Debug().Str("socketID", u.SocketID).Str("session", sess.SessionID).Msg("removing user from session")
					}
				}
				sess.Participants = newUsers
				return errors.WithStack(saveSession(ctx, *sess))
			})
			if err != nil {
				return errors.WithStack(err)
			}
		}
		return nil
	}
}

type InterestRecorder func(ctx context.Context, sessionID string, connectionID string) error

func NewInterestRecorder(dynamo DynamoClient, sessionTable string, watcherTable string, locker lock.GlobalLockAppropriator, sessionExpiration time.Duration) InterestRecorder {
	return func(ctx context.Context, sessionID string, connectionID string) error {
		expiration := time.Now().Add(sessionExpiration * 10)
		err := recordWatcher(ctx, dynamo, expiration, sessionTable, "ConnectionID", "Sessions", connectionID, sessionID)
		if err != nil {
			return errors.WithStack(err)
		}
		return locker.DoWithLock(ctx, lock.SessionInterestLockKey(sessionID), func(ctx context.Context) error {
			return errors.WithStack(recordWatcher(ctx, dynamo, expiration, watcherTable, "SessionID", "Connections", sessionID, connectionID))
		})
	}
}

func recordWatcher(ctx context.Context, dynamo DynamoClient, expiration time.Time, table, key, valueKey, observer, subject string) error {
	existingRes, err := dynamo.GetItemWithContext(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(table),
		Key: map[string]*dynamodb.AttributeValue{
			key: {S: aws.String(observer)},
		},
	})
	if err != nil {
		return errors.WithStack(err)
	}
	_, err = dynamo.PutItemWithContext(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(table),
		Item: map[string]*dynamodb.AttributeValue{
			key:      {S: aws.String(observer)},
			valueKey: {L: watchList(existingRes, key, valueKey, subject)},
			// TODO - this sucks and we should update expirations on session changes but -meh- for now
			"Expiration": {N: aws.String(strconv.FormatInt(expiration.Unix(), 10))},
		},
	})
	return errors.WithStack(err)
}

func watchList(current *dynamodb.GetItemOutput, key, valueKey, newValue string) []*dynamodb.AttributeValue {
	if current.Item[key] == nil {
		return []*dynamodb.AttributeValue{{S: aws.String(newValue)}}
	}
	ret := make([]*dynamodb.AttributeValue, len(current.Item[valueKey].L))
	alreadyPresent := false
	for i, r := range current.Item[valueKey].L {
		if *r.S == newValue {
			alreadyPresent = true
		}
		ret[i] = r
	}
	if !alreadyPresent {
		ret = append(ret, &dynamodb.AttributeValue{S: aws.String(newValue)})
	}
	return ret
}
