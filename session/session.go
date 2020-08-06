package session

import (
	"context"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"strconv"
	"time"
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

func NewStarter(dynamo *dynamodb.DynamoDB, tableName string, sessionExpiration time.Duration) Starter {
	return func(ctx context.Context, toStart StartRequest) (CompleteSessionView, error) {
		sessionID := uuid.New().String()
		facilitatorSessionKey := uuid.New().String()
		toPut := &dynamodb.PutItemInput{
			TableName: aws.String(tableName),
			Item: map[string]*dynamodb.AttributeValue{
				"SessionID":             {S: aws.String(sessionID)},
				"VotesShown":            {BOOL: aws.Bool(false)},
				"FacilitatorSessionKey": {S: aws.String(facilitatorSessionKey)},
				"Facilitator":           {M: convertUser(toStart.Facilitator)},
				"FacilitatorPoints":     {BOOL: aws.Bool(toStart.FacilitatorPoints)},
				"Participants":          {L: []*dynamodb.AttributeValue{}},
				"Expiration":            {N: aws.String(strconv.FormatInt(time.Now().Add(sessionExpiration).Unix(), 10))},
			},
			ConditionExpression: aws.String("attribute_not_exists(LockID)"),
		}

		_, err := dynamo.PutItemWithContext(ctx, toPut)
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

func NewSaver(dynamo *dynamodb.DynamoDB, tableName string, notifyObservers ChangeNotifier, sessionExpiration time.Duration) Saver {
	return func(ctx context.Context, toSave CompleteSessionView) error {
		participants := make([]*dynamodb.AttributeValue, len(toSave.Participants))
		for i, u := range toSave.Participants {
			participants[i] = &dynamodb.AttributeValue{M: convertUser(u)}
		}
		toPut := &dynamodb.PutItemInput{
			TableName: aws.String(tableName),
			Item: map[string]*dynamodb.AttributeValue{
				"SessionID":             {S: aws.String(toSave.SessionID)},
				"VotesShown":            {BOOL: aws.Bool(toSave.VotesShown)},
				"FacilitatorSessionKey": {S: aws.String(toSave.FacilitatorSessionKey)},
				"Facilitator":           {M: convertUser(toSave.Facilitator)},
				"FacilitatorPoints":     {BOOL: aws.Bool(toSave.FacilitatorPoints)},
				"Participants":          {L: participants},
				"Expiration":            {N: aws.String(strconv.FormatInt(time.Now().Add(sessionExpiration).Unix(), 10))},
			},
			ConditionExpression: aws.String("attribute_not_exists(LockID)"),
		}

		_, err := dynamo.PutItemWithContext(ctx, toPut)
		if err != nil {
			return errors.WithStack(err)
		}
		return errors.WithStack(notifyObservers(ctx, toSave))
	}
}

type Loader func(ctx context.Context, sessionID string) (*CompleteSessionView, error)

func NewLoader(dynamo *dynamodb.DynamoDB, tableName string) Loader {
	return func(ctx context.Context, sessionID string) (*CompleteSessionView, error) {
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
		return &CompleteSessionView{
			SessionID:             *res.Item["SessionID"].S,
			VotesShown:            *res.Item["VotesShown"].BOOL,
			FacilitatorSessionKey: *res.Item["FacilitatorSessionKey"].S,
			Facilitator:           readUser(res.Item["Facilitator"].M),
			FacilitatorPoints:     *res.Item["FacilitatorPoints"].BOOL,
			Participants:          participants,
		}, nil
	}
}

func convertUser(u User) map[string]*dynamodb.AttributeValue {
	ret := map[string]*dynamodb.AttributeValue{
		"UserID":   {S: aws.String(u.UserID)},
		"Name":     {S: aws.String(u.Name)},
		"Handle":   {S: aws.String(u.Handle)},
		"SocketID": {S: aws.String(u.SocketID)},
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
