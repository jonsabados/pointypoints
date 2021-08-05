package profile

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/pkg/errors"
)

const (
	fieldUserID = "UserID"
	fieldEmail  = "Email"
	fieldName   = "UserName" // Name is reserved
	fieldHandle = "Handle"

	fieldSessionStartCount = "SessionStartCount"
	fieldSessionWatchCount = "SessionWatchCount"
	fieldSessionJoinCount  = "SessionJoinCount"
	fieldVoteCount         = "VoteCount"
)

type Profile struct {
	UserID string
	Email  string
	Name   string
	Handle *string
}

type UserView struct {
	Email  string  `json:"email"`
	Name   string  `json:"name"`
	Handle *string `json:"handle"`
}

type Fetcher func(ctx context.Context, userID string) (*Profile, error)

func NewFetcher(dynamo *dynamodb.DynamoDB, tableName string) Fetcher {
	return func(ctx context.Context, userID string) (*Profile, error) {
		res, err := dynamo.GetItemWithContext(ctx, &dynamodb.GetItemInput{
			TableName: aws.String(tableName),
			Key: map[string]*dynamodb.AttributeValue{
				"UserID": {S: aws.String(userID)},
			},
			ProjectionExpression: aws.String(fmt.Sprintf("%s,%s,%s", fieldName, fieldEmail, fieldHandle)),
		})

		if err != nil {
			return nil, errors.Wrap(err, "error reading user from dynamo")
		}

		if res.Item == nil {
			return nil, nil
		}

		ret := &Profile{
			UserID: userID,
			Email:  *res.Item[fieldEmail].S,
			Name:   *res.Item[fieldName].S,
		}

		if i, ok := res.Item[fieldHandle]; ok {
			ret.Handle = i.S
		}

		return ret, nil
	}
}

type Writer func(ctx context.Context, profile Profile) error

func NewWriter(dynamo *dynamodb.DynamoDB, tableName string) Writer {
	return func(ctx context.Context, profile Profile) error {
		item := map[string]*dynamodb.AttributeValue{
			fieldUserID: {S: aws.String(profile.UserID)},
			fieldName:   {S: aws.String(profile.Name)},
			fieldEmail:  {S: aws.String(profile.Email)},
		}

		if profile.Handle != nil {
			item[fieldHandle] = &dynamodb.AttributeValue{S: profile.Handle}
		}

		_, err := dynamo.PutItemWithContext(ctx, &dynamodb.PutItemInput{
			TableName: aws.String(tableName),
			Item:      item,
		})
		return errors.Wrap(err, "error writing profile")
	}
}

type StatsUpdateFactory struct {
	tableName string
}

func (s *StatsUpdateFactory) SessionIncrement(userID string) *dynamodb.Update {
	return s.statsColumnIncrement(userID, fieldSessionStartCount)
}

func (s *StatsUpdateFactory) SessionWatchIncrement(userID string) *dynamodb.Update {
	return s.statsColumnIncrement(userID, fieldSessionWatchCount)
}

func (s *StatsUpdateFactory) SessionJoinIncrement(userID string) *dynamodb.Update {
	return s.statsColumnIncrement(userID, fieldSessionJoinCount)
}

func (s *StatsUpdateFactory) VoteIncrement(userID string) *dynamodb.Update {
	return s.statsColumnIncrement(userID, fieldVoteCount)
}


func (s *StatsUpdateFactory) statsColumnIncrement(userID string, column string) *dynamodb.Update {
	return &dynamodb.Update{
		TableName: aws.String(s.tableName),
		Key: map[string]*dynamodb.AttributeValue{
			"UserID": {S: aws.String(userID)},
		},
		UpdateExpression: aws.String(fmt.Sprintf("ADD %s :inc", column)),
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{":inc": {N: aws.String("1")}},
	}
}

func NewStatsUpdateFactory(profileTable string) *StatsUpdateFactory {
	return &StatsUpdateFactory{profileTable}
}
