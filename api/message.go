package api

import (
	"context"
	"encoding/json"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/apigatewaymanagementapi"
	"github.com/pkg/errors"
)

type MessageType string

const (
	SessionCreated           = MessageType("SESSION_CREATED")
	SessionUpdated           = MessageType("SESSION_UPDATED")
	FacilitatorSessionLoaded = MessageType("FACILITATOR_SESSION_LOADED")
	SessionLoaded            = MessageType("SESSION_LOADED")
	ErrorEncountered         = MessageType("ERROR_ENCOUNTERED")
)

type ConnectionPoster interface {
	PostToConnectionWithContext(ctx aws.Context, input *apigatewaymanagementapi.PostToConnectionInput, opts ...request.Option) (*apigatewaymanagementapi.PostToConnectionOutput, error)
}

type Message struct {
	Type MessageType `json:"type"`
	Body interface{} `json:"body"`
}

type MessageDispatcher func(ctx context.Context, connectionID string, message Message) error

func NewMessageDispatcher(gateway ConnectionPoster) MessageDispatcher {
	return func(ctx context.Context, connectionID string, message Message) error {
		body, err := json.Marshal(message)
		if err != nil {
			return errors.WithStack(err)
		}
		_, err = gateway.PostToConnectionWithContext(ctx, &apigatewaymanagementapi.PostToConnectionInput{
			ConnectionId: aws.String(connectionID),
			Data:         body,
		})
		return errors.WithStack(err)
	}
}
