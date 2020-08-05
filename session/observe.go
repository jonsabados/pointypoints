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

func NewChangeNotifier(dynamo *dynamodb.DynamoDB, watcherTable string, dispatchMessage api.MessageDispatcher) ChangeNotifier {
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
				Body: updated,
			})
			if err != nil {
				zerolog.Ctx(ctx).Warn().Str("error", fmt.Sprintf("%+v", err)).Msg("error notifying observer")
			}
		}
		return nil
	}
}

type InterestRecorder func(ctx context.Context, sessionID string, connectionID string) error

func NewInterestRecorder(dynamo *dynamodb.DynamoDB, sessionTable string, watcherTable string, locker lock.GlobalLockAppropriator, sessionExpiration time.Duration) InterestRecorder {
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

func recordWatcher(ctx context.Context, dynamo *dynamodb.DynamoDB, expiration time.Time, table, key, valueKey, observer, subject string) error {
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
