package session

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
)

type UserType int

const (
	Facilitator UserType = iota
	Participant
)

const (
	sessionRecordRangeKeyValue      = "session"
	facilitatorRecordRangeKeyValue  = "facilitator"
	participantRecordRangeKeyPrefix = "user:"
	watcherRecordRangeKeyPrefix     = "watcher:"
)

type User struct {
	UserID      string  `json:"userId"`
	Name        string  `json:"name,omitempty"`
	Handle      string  `json:"handle,omitempty"`
	CurrentVote *string `json:"currentVote,omitempty"`
	SocketID    string  `json:"-"`
}

type StartRequest struct {
	Facilitator       User `json:"facilitator"`
	FacilitatorPoints bool `json:"facilitatorPoints"`
}

type LoadFacilitatorSessionRequest struct {
	SessionID             string `json:"sessionId"`
	FacilitatorSessionKey string `json:"facilitatorSessionKey"`
	MarkActive            bool   `json:"markActive"`
}

type LoadSessionRequest struct {
	SessionID string `json:"sessionId"`
}

type JoinSessionRequest struct {
	SessionID string `json:"sessionId"`
	User      User   `json:"user"`
}

type VoteRequest struct {
	SessionID string `json:"sessionId"`
	Vote      string `json:"vote"`
}

type ShowVotesRequest struct {
	SessionID             string `json:"sessionId"`
	FacilitatorSessionKey string `json:"facilitatorSessionKey"`
}

type CompleteSessionView struct {
	SessionID             string `json:"sessionId"`
	VotesShown            bool   `json:"votesShown"`
	FacilitatorSessionKey string `json:"facilitatorSessionKey,omitempty"`
	Facilitator           User   `json:"facilitator"`
	FacilitatorPoints     bool   `json:"facilitatorPoints"`
	Participants          []User `json:"participants"`
}

type ParticipantSessionView struct {
	SessionID         string `json:"sessionId"`
	VotesShown        bool   `json:"votesShown"`
	Facilitator       User   `json:"facilitator"`
	FacilitatorPoints bool   `json:"facilitatorPoints"`
	Participants      []User `json:"participants"`
}

type DynamoClient interface {
	GetItemWithContext(ctx aws.Context, input *dynamodb.GetItemInput, opts ...request.Option) (*dynamodb.GetItemOutput, error)
	QueryWithContext(ctx aws.Context, input *dynamodb.QueryInput, opts ...request.Option) (*dynamodb.QueryOutput, error)
	PutItemWithContext(ctx aws.Context, input *dynamodb.PutItemInput, opts ...request.Option) (*dynamodb.PutItemOutput, error)
	TransactWriteItemsWithContext(ctx aws.Context, input *dynamodb.TransactWriteItemsInput, opts ...request.Option) (*dynamodb.TransactWriteItemsOutput, error)
	DeleteItemWithContext(ctx aws.Context, input *dynamodb.DeleteItemInput, opts ...request.Option) (*dynamodb.DeleteItemOutput, error)
}

func ToParticipantView(s CompleteSessionView, connectionID string) ParticipantSessionView {
	participants := make([]User, len(s.Participants))
	for i, u := range s.Participants {
		participants[i] = participantUserView(s, u, connectionID)
	}
	return ParticipantSessionView{
		SessionID:         s.SessionID,
		VotesShown:        s.VotesShown,
		Facilitator:       participantUserView(s, s.Facilitator, connectionID),
		FacilitatorPoints: s.FacilitatorPoints,
		Participants:      participants,
	}
}

func participantUserView(s CompleteSessionView, u User, connectionID string) User {
	ret := User{
		UserID: u.UserID,
		Handle: u.Handle,
	}
	if u.Handle == "" {
		ret.Name = u.Name
	}
	if s.VotesShown || u.SocketID == connectionID {
		ret.CurrentVote = u.CurrentVote
	}
	return ret
}

type Starter func(ctx context.Context, toStart StartRequest) (CompleteSessionView, error)

func NewStarter(dynamo DynamoClient, tableName string, sessionExpiration time.Duration) Starter {
	return func(ctx context.Context, toStart StartRequest) (CompleteSessionView, error) {
		sessionID := uuid.New().String()
		facilitatorSessionKey := uuid.New().String()

		expiration := &dynamodb.AttributeValue{N: aws.String(strconv.FormatInt(time.Now().Add(sessionExpiration).Unix(), 10))}

		sessionPut := &dynamodb.Put{
			TableName: aws.String(tableName),
			Item: map[string]*dynamodb.AttributeValue{
				"SessionID":             {S: aws.String(sessionID)},
				"RangeKey":              {S: aws.String(sessionRecordRangeKeyValue)},
				"VotesShown":            {BOOL: aws.Bool(false)},
				"FacilitatorSessionKey": {S: aws.String(facilitatorSessionKey)},
				"FacilitatorPoints":     {BOOL: aws.Bool(toStart.FacilitatorPoints)},
				"Expiration":            expiration,
			},
		}

		facilitatorPut := &dynamodb.Put{
			TableName: aws.String(tableName),
			Item:      convertUser(sessionID, Facilitator, toStart.Facilitator, expiration),
		}

		_, err := dynamo.TransactWriteItemsWithContext(ctx, &dynamodb.TransactWriteItemsInput{
			TransactItems: []*dynamodb.TransactWriteItem{
				{
					Put: sessionPut,
				},
				{
					Put: facilitatorPut,
				},
			},
		})
		return CompleteSessionView{
			SessionID:             sessionID,
			FacilitatorSessionKey: facilitatorSessionKey,
			Facilitator:           toStart.Facilitator,
			FacilitatorPoints:     toStart.FacilitatorPoints,
			Participants:          make([]User, 0),
		}, errors.WithStack(err)
	}
}

type Saver func(ctx context.Context, toSave CompleteSessionView) error

func NewSaver(dynamo DynamoClient, tableName string, notifyObservers ChangeNotifier, sessionExpiration time.Duration) Saver {
	return func(ctx context.Context, toSave CompleteSessionView) error {
		expiration := &dynamodb.AttributeValue{N: aws.String(strconv.FormatInt(time.Now().Add(sessionExpiration).Unix(), 10))}

		sessionPut := &dynamodb.Put{
			TableName: aws.String(tableName),
			Item: map[string]*dynamodb.AttributeValue{
				"SessionID":             {S: aws.String(toSave.SessionID)},
				"RangeKey":              {S: aws.String(sessionRecordRangeKeyValue)},
				"VotesShown":            {BOOL: aws.Bool(toSave.VotesShown)},
				"FacilitatorSessionKey": {S: aws.String(toSave.FacilitatorSessionKey)},
				"FacilitatorPoints":     {BOOL: aws.Bool(toSave.FacilitatorPoints)},
				"Expiration":            {N: aws.String(strconv.FormatInt(time.Now().Add(sessionExpiration).Unix(), 10))},
			},
		}

		facilitatorPut := &dynamodb.Put{
			TableName: aws.String(tableName),
			Item:      convertUser(toSave.SessionID, Facilitator, toSave.Facilitator, expiration),
		}

		transactItems := []*dynamodb.TransactWriteItem{
			{
				Put: sessionPut,
			},
			{
				Put: facilitatorPut,
			},
		}

		for _, u := range toSave.Participants {
			transactItems = append(transactItems, &dynamodb.TransactWriteItem{
				Put: &dynamodb.Put{
					TableName: aws.String(tableName),
					Item:      convertUser(toSave.SessionID, Participant, u, expiration),
				},
			})
		}

		_, err := dynamo.TransactWriteItemsWithContext(ctx, &dynamodb.TransactWriteItemsInput{
			TransactItems: transactItems,
		})
		if err != nil {
			return errors.WithStack(err)
		}
		return errors.WithStack(notifyObservers(ctx, toSave))
	}
}

type UserSaver func(ctx context.Context, sessionID string, user User, userType UserType) error

func NewUserSaver(dynamo DynamoClient, tableName string, sessionExpiration time.Duration) UserSaver {
	return func(ctx context.Context, sessionID string, user User, userType UserType) error {
		expiration := &dynamodb.AttributeValue{N: aws.String(strconv.FormatInt(time.Now().Add(sessionExpiration).Unix(), 10))}

		// swap watcher for user
		_, err := dynamo.DeleteItemWithContext(ctx, &dynamodb.DeleteItemInput{
			TableName: aws.String(tableName),
			Key: map[string]*dynamodb.AttributeValue{
				"SessionID": {S: aws.String(sessionID)},
				"RangeKey":  {S: aws.String(fmt.Sprintf("%s%s", watcherRecordRangeKeyPrefix, user.SocketID))},
			},
		})
		if err != nil {
			return errors.WithStack(err)
		}

		_, err = dynamo.PutItemWithContext(ctx, &dynamodb.PutItemInput{
			TableName: aws.String(tableName),
			Item:      convertUser(sessionID, userType, user, expiration),
		})

		return errors.WithStack(err)
	}
}

type Loader func(ctx context.Context, sessionID string) (*CompleteSessionView, error)

func NewLoader(dynamo DynamoClient, tableName string) Loader {
	return func(ctx context.Context, sessionID string) (*CompleteSessionView, error) {
		res, err := dynamo.QueryWithContext(ctx, &dynamodb.QueryInput{
			TableName: aws.String(tableName),
			KeyConditions: map[string]*dynamodb.Condition{
				"SessionID": {
					ComparisonOperator: aws.String("EQ"),
					AttributeValueList: []*dynamodb.AttributeValue{
						{S: aws.String(sessionID)},
					},
				},
			},
		})

		if err != nil {
			return nil, errors.WithStack(err)
		}

		if *res.Count == 0 {
			return nil, nil
		}

		// if there are no participants we still want a non-nil list since every other language, including javascript
		// will grenade if it works with a null list
		ret := &CompleteSessionView{
			Participants: make([]User, 0),
		}
		for _, item := range res.Items {
			rangeKey := *item["RangeKey"].S
			if rangeKey == sessionRecordRangeKeyValue {
				ret.SessionID = *item["SessionID"].S
				ret.VotesShown = *item["VotesShown"].BOOL
				ret.FacilitatorSessionKey = *item["FacilitatorSessionKey"].S
				ret.FacilitatorPoints = *item["FacilitatorPoints"].BOOL
			} else if rangeKey == facilitatorRecordRangeKeyValue {
				ret.Facilitator = readUser(item)
			} else if strings.HasPrefix(rangeKey, participantRecordRangeKeyPrefix) {
				ret.Participants = append(ret.Participants, readUser(item))
			} else if !strings.HasPrefix(rangeKey, watcherRecordRangeKeyPrefix) {
				zerolog.Ctx(ctx).Warn().Interface("record", item).Msg("unexpected record spotted")
			}
		}

		return ret, nil
	}
}

func userRangeKey(user User, userType UserType) *dynamodb.AttributeValue {
	switch userType {
	case Facilitator:
		return &dynamodb.AttributeValue{S: aws.String(facilitatorRecordRangeKeyValue)}
	case Participant:
		return &dynamodb.AttributeValue{S: aws.String(fmt.Sprintf("%s%s", participantRecordRangeKeyPrefix, user.SocketID))}
	default:
		panic(fmt.Sprintf("unknown user type %d", userType))
	}
}

func convertUser(sessionID string, userType UserType, u User, expiration *dynamodb.AttributeValue) map[string]*dynamodb.AttributeValue {
	ret := map[string]*dynamodb.AttributeValue{
		"SessionID":  {S: aws.String(sessionID)},
		"RangeKey":   userRangeKey(u, userType),
		"UserID":     {S: aws.String(u.UserID)},
		"Name":       {S: aws.String(u.Name)},
		"Handle":     {S: aws.String(u.Handle)},
		"SocketID":   {S: aws.String(u.SocketID)},
		"Expiration": expiration,
	}
	if u.CurrentVote != nil {
		ret["CurrentVote"] = &dynamodb.AttributeValue{S: u.CurrentVote}
	}
	return ret
}

func readUser(r map[string]*dynamodb.AttributeValue) User {
	ret := User{
		UserID:   *r["UserID"].S,
		Name:     *r["Name"].S,
		Handle:   *r["Handle"].S,
		SocketID: *r["SocketID"].S,
	}
	if r["CurrentVote"] != nil {
		ret.CurrentVote = r["CurrentVote"].S
	}
	return ret
}
