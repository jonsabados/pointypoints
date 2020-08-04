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
	Name     string `json:"name,omitempty"`
	Handle   string `json:"handle,omitempty"`
	SocketID string `json:"-"`
}

type StartRequest struct {
	Facilitator       User `json:"facilitator"`
	FacilitatorPoints bool `json:"facilitatorPoints"`
}

type FacilitatorSessionVew struct {
	SessionID             string `json:"sessionId"`
	FacilitatorSessionKey string `json:"facilitatorSessionKey,omitempty"`
	Facilitator           User   `json:"facilitator"`
	FacilitatorPoints     bool   `json:"facilitatorPoints"`
	Participants          []User `json:"participants"`
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

func convertUser(u User) map[string]*dynamodb.AttributeValue {
	return map[string]*dynamodb.AttributeValue{
		"Name":   {S: aws.String(u.Name)},
		"Handle": {S: aws.String(u.Handle)},
	}
}
