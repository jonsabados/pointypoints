package session

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/jonsabados/goauth"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"

	"github.com/jonsabados/pointypoints/api"
	"github.com/jonsabados/pointypoints/profile"
)

type ChangeNotifier func(ctx context.Context, updated CompleteSessionView) error

func NewChangeNotifier(dynamo DynamoClient, tableName string, dispatchMessage api.MessageDispatcher) ChangeNotifier {
	return func(ctx context.Context, updated CompleteSessionView) error {
		records, err := dynamo.QueryWithContext(ctx, &dynamodb.QueryInput{
			TableName: aws.String(tableName),
			KeyConditions: map[string]*dynamodb.Condition{
				"SessionID": {
					ComparisonOperator: aws.String("EQ"),
					AttributeValueList: []*dynamodb.AttributeValue{
						{S: aws.String(updated.SessionID)},
					},
				},
			},
		})
		if err != nil {
			return errors.WithStack(err)
		}
		for _, r := range records.Items {
			if socketID, ok := r["SocketID"]; ok {
				err := dispatchMessage(ctx, *socketID.S, api.Message{
					Type: api.SessionUpdated,
					Body: connectionView(updated, *socketID.S),
				})
				if err != nil {
					zerolog.Ctx(ctx).Warn().Err(err).Msg("error notifying observer")
				}
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

type WatcherSaver func(ctx context.Context, initiator goauth.Principal, sessionID string, socketID string) error

func NewWatcherSaver(dynamo DynamoClient, tableName string, sessionExpiration time.Duration,  sf *profile.StatsUpdateFactory) WatcherSaver {
	return func(ctx context.Context, initiator goauth.Principal, sessionID string, socketID string) error {
		expiration := &dynamodb.AttributeValue{N: aws.String(strconv.FormatInt(time.Now().Add(sessionExpiration).Unix(), 10))}

		watchPut := &dynamodb.Put{
			TableName: aws.String(tableName),
			Item: map[string]*dynamodb.AttributeValue{
				"SessionID":  {S: aws.String(sessionID)},
				"RangeKey":   {S: aws.String(fmt.Sprintf("%s%s", watcherRecordRangeKeyPrefix, socketID))},
				"SocketID":   {S: aws.String(socketID)},
				"Expiration": expiration,
			},
		}

		_, err := dynamo.TransactWriteItemsWithContext(ctx, &dynamodb.TransactWriteItemsInput{
			TransactItems: []*dynamodb.TransactWriteItem{
				{
					Put: watchPut,
				},
				{
					Update: sf.SessionWatchIncrement(initiator.UserID),
				},
			},
		})

		return errors.WithStack(err)
	}
}

type Disconnector func(ctx context.Context, connectionID string) error

func NewDisconnector(dynamo DynamoClient, tableName string, indexName string, loadSession Loader, notifyParticipants ChangeNotifier) Disconnector {
	return func(ctx context.Context, connectionID string) error {
		records, err := dynamo.QueryWithContext(ctx, &dynamodb.QueryInput{
			TableName: aws.String(tableName),
			IndexName: aws.String(indexName),
			KeyConditions: map[string]*dynamodb.Condition{
				"SocketID": {
					ComparisonOperator: aws.String("EQ"),
					AttributeValueList: []*dynamodb.AttributeValue{
						{S: aws.String(connectionID)},
					},
				},
			},
		})
		if err != nil {
			return errors.WithStack(err)
		}

		for _, r := range records.Items {
			_, err := dynamo.DeleteItemWithContext(ctx, &dynamodb.DeleteItemInput{
				TableName: aws.String(tableName),
				Key: map[string]*dynamodb.AttributeValue{
					"SessionID": r["SessionID"],
					"RangeKey":  r["RangeKey"],
				},
			})
			if err != nil {
				return errors.WithStack(err)
			}

			sess, err := loadSession(ctx, *r["SessionID"].S)
			if err != nil {
				return errors.WithStack(err)
			}

			err = notifyParticipants(ctx, *sess)
			if err != nil {
				return errors.WithStack(err)
			}
		}
		return nil
	}
}

