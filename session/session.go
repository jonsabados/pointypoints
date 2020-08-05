package session

import (
	"context"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/google/uuid"
	"github.com/jonsabados/pointypoints/lock"
	"github.com/pkg/errors"
	"strconv"
	"time"
)

type User struct {
	Name        string `json:"name,omitempty"`
	Handle      string `json:"handle,omitempty"`
	CurrentVote *int   `json:"currentVote,omitempty"`
	SocketID    string `json:"-"`
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

type FacilitatorSessionVew struct {
	SessionID             string `json:"sessionId"`
	FacilitatorSessionKey string `json:"facilitatorSessionKey,omitempty"`
	Facilitator           User   `json:"facilitator"`
	FacilitatorPoints     bool   `json:"facilitatorPoints"`
	Participants          []User `json:"participants"`
}

type ParticipantSessionView struct {
	SessionID         string `json:"sessionId"`
	Facilitator       User   `json:"facilitator"`
	FacilitatorPoints bool   `json:"facilitatorPoints"`
	Participants      []User `json:"participants"`
}

func ToParticipantView(s FacilitatorSessionVew) ParticipantSessionView {
	participants := make([]User, len(s.Participants))
	for i, u := range s.Participants {
		participants[i] = ParticipantUserView(u)
	}
	return ParticipantSessionView{
		SessionID:         s.SessionID,
		Facilitator:       ParticipantUserView(s.Facilitator),
		FacilitatorPoints: s.FacilitatorPoints,
		Participants:      participants,
	}
}

func ParticipantUserView(u User) User {
	ret := User{
		Handle: u.Handle,
	}
	if u.Handle == "" {
		ret.Name = u.Name
	}
	return ret
}

type Starter func(ctx context.Context, toStart StartRequest) (FacilitatorSessionVew, error)

func NewStarter(dynamo *dynamodb.DynamoDB, tableName string, sessionExpiration time.Duration) Starter {
	return func(ctx context.Context, toStart StartRequest) (FacilitatorSessionVew, error) {
		sessionID := uuid.New().String()
		facilitatorSessionKey := uuid.New().String()
		toPut := &dynamodb.PutItemInput{
			TableName: aws.String(tableName),
			Item: map[string]*dynamodb.AttributeValue{
				"SessionID":             {S: aws.String(sessionID)},
				"FacilitatorSessionKey": {S: aws.String(facilitatorSessionKey)},
				"Facilitator":           {M: convertUser(toStart.Facilitator)},
				"FacilitatorPoints":     {BOOL: aws.Bool(toStart.FacilitatorPoints)},
				"Participants":          {L: []*dynamodb.AttributeValue{}},
				"Expiration":            {N: aws.String(strconv.FormatInt(time.Now().Add(sessionExpiration).Unix(), 10))},
			},
			ConditionExpression: aws.String("attribute_not_exists(LockID)"),
		}

		_, err := dynamo.PutItemWithContext(ctx, toPut)
		return FacilitatorSessionVew{
			SessionID:             sessionID,
			FacilitatorSessionKey: facilitatorSessionKey,
			Facilitator:           toStart.Facilitator,
			FacilitatorPoints:     toStart.FacilitatorPoints,
			Participants:          make([]User, 0),
		}, errors.WithStack(err)
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

type Loader func(ctx context.Context, sessionID string) (*FacilitatorSessionVew, error)

func NewLoader(dynamo *dynamodb.DynamoDB, tableName string) Loader {
	return func(ctx context.Context, sessionID string) (*FacilitatorSessionVew, error) {
		res, err := dynamo.GetItemWithContext(ctx, &dynamodb.GetItemInput{
			TableName: aws.String(tableName),
			Key: map[string]*dynamodb.AttributeValue{
				"SessionID": {S: aws.String(sessionID)},
			},
		})
		if err != nil {
			return nil, errors.WithStack(err)
		}
		if res.Item["SessionID"].S == nil {
			return nil, nil
		}
		rawParticipants := res.Item["Participants"].L
		participants := make([]User, len(rawParticipants))
		for i, r := range rawParticipants {
			participants[i] = readUser(r.M)
		}
		return &FacilitatorSessionVew{
			SessionID:             *res.Item["SessionID"].S,
			FacilitatorSessionKey: *res.Item["FacilitatorSessionKey"].S,
			Facilitator:           readUser(res.Item["Facilitator"].M),
			FacilitatorPoints:     *res.Item["FacilitatorPoints"].BOOL,
			Participants:          participants,
		}, nil
	}
}

func convertUser(u User) map[string]*dynamodb.AttributeValue {
	ret := map[string]*dynamodb.AttributeValue{
		"Name":   {S: aws.String(u.Name)},
		"Handle": {S: aws.String(u.Handle)},
	}
	if u.CurrentVote != nil {
		ret["CurrentVote"] = &dynamodb.AttributeValue{N: aws.String(strconv.Itoa(*u.CurrentVote))}
	}
	return ret
}

func readUser(r map[string]*dynamodb.AttributeValue) User {
	ret := User{
		Name:   *r["Name"].S,
		Handle: *r["Handle"].S,
	}
	if r["CurrentVote"] != nil {
		val, err := strconv.Atoi(*r["CurrentVote"].N)
		if err != nil {
			panic(err)
		}
		ret.CurrentVote = &val
	}
	return ret
}
