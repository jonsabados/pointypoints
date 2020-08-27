package lock

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"strconv"
	"time"
)

// Lock represents a distributed lock
type Lock interface {
	// Unlock releases the lock
	Unlock(ctx context.Context) error
}

type lockImpl struct {
	lockID     string
	tableName  string
	dynamo     *dynamodb.DynamoDB
}

func (l *lockImpl) Unlock(ctx context.Context) error {
	input := &dynamodb.DeleteItemInput{
		TableName: aws.String(l.tableName),
		Key: map[string]*dynamodb.AttributeValue{
			"LockID": {S: aws.String(l.lockID)},
		},
	}
	_, err := l.dynamo.DeleteItemWithContext(ctx, input)
	return errors.WithStack(err)
}

// GlobalLockAppropriator can be used to acquire global locks on resources
type GlobalLockAppropriator func(ctx context.Context, lockID string) (Lock, error)

// NewGlobalLockAppropriator returns a fully wired GlobalLockAppropriator. If lock acquisition fails it will be retried
// based on the given duration. Lock entries will expire in dynamo based on dynamoExpiration. Context expiration is respected.
func NewGlobalLockAppropriator(dynamo *dynamodb.DynamoDB, tableName string, retry time.Duration, dynamoExpiration time.Duration) GlobalLockAppropriator {
	return func(ctx context.Context, lockID string) (Lock, error) {
		expiration := time.Now().Add(dynamoExpiration).Unix()
		lockAcquired := false
		for !lockAcquired {
			select {
			case <- ctx.Done():
				return nil, errors.New("context closed")
			default:
				toPut := &dynamodb.PutItemInput{
					TableName: aws.String(tableName),
					Item: map[string]*dynamodb.AttributeValue{
						"LockID":     {S: aws.String(lockID)},
						"Expiration": {N: aws.String(strconv.FormatInt(expiration, 10))},
					},
					ConditionExpression: aws.String("attribute_not_exists(LockID)"),
				}

				_, err := dynamo.PutItemWithContext(ctx, toPut)
				if err != nil {
					zerolog.Ctx(ctx).Debug().Err(err).Msg("waiting for lock")
					time.Sleep(retry)
					continue
				}
				lockAcquired = true
			}
		}
		return &lockImpl{
			lockID:     lockID,
			dynamo:     dynamo,
			tableName:  tableName,
		}, nil
	}
}

func (a GlobalLockAppropriator) DoWithLock(ctx context.Context, lockID string, action func(ctx context.Context) error) error {
	lock, err := a(ctx, lockID)
	if err != nil {
		return errors.WithStack(err)
	}
	defer func() {
		err := lock.Unlock(ctx)
		if err != nil {
			zerolog.Ctx(ctx).Error().Str("error", fmt.Sprintf("%+v", err)).Msg("unable to release lock")
		}
	}()
	return errors.WithStack(action(ctx))
}

func SessionLockKey(sessionID string) string {
	return fmt.Sprintf("session:%s", sessionID)
}

func SessionInterestLockKey(sessionID string) string {
	return fmt.Sprintf("sessionInterest:%s", sessionID)
}