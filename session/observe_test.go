package session

import (
	"context"
	"errors"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/jonsabados/pointypoints/api"
	"github.com/jonsabados/pointypoints/session/testutil"
	"github.com/stretchr/testify/assert"
	"testing"
)

var emptyOpts []request.Option

func Test_NewChangeNotifier_ErrorLoadingWatchers(t *testing.T) {
	asserter := assert.New(t)

	inputCtx := testutil.NewTestContext()

	dynamo := &testutil.MockDynamoClient{}
	tableName := "watchers"

	sessionID := "abcdefg"

	facilitatorConnectionID := "facilitator"
	facilitatorSessionKey := "bobsuruncle"
	userAConnectionID := "aaaaaaaa"
	userBConnectionID := "bbbbbbbb"

	input := CompleteSessionView{
		SessionID:             sessionID,
		VotesShown:            true,
		FacilitatorSessionKey: facilitatorSessionKey,
		Facilitator: User{
			UserID:      "someUUIDFacilitator",
			Name:        "Bob",
			Handle:      "TheTester",
			CurrentVote: nil,
			SocketID:    facilitatorConnectionID,
		},
		FacilitatorPoints: true,
		Participants: []User{
			{
				UserID:      "someUUIDA",
				Name:        "A",
				Handle:      "AAA",
				CurrentVote: aws.String("1"),
				SocketID:    userAConnectionID,
			},
			{
				UserID:      "someUUIDB",
				Name:        "B",
				Handle:      "BBB",
				CurrentVote: aws.String("B"),
				SocketID:    userBConnectionID,
			},
		},
	}
	expectedError := "kablam"

	dispatcher := api.MessageDispatcher(func(ctx context.Context, connectionID string, message api.Message) error {
		asserter.Fail("nothing should have been dispatched")
		return nil
	})

	dynamo.On("QueryWithContext", inputCtx, &dynamodb.QueryInput{
		TableName: aws.String(tableName),
		KeyConditions: map[string]*dynamodb.Condition{
			"SessionID": {
				ComparisonOperator: aws.String("EQ"),
				AttributeValueList: []*dynamodb.AttributeValue{
					{S: aws.String(sessionID)},
				},
			},
		},
	}, emptyOpts).Return(nil, errors.New(expectedError))

	err := NewChangeNotifier(dynamo, tableName, dispatcher)(inputCtx, input)
	asserter.EqualError(err, expectedError)
}

func Test_NewChangeNotifier_VotesShown(t *testing.T) {
	asserter := assert.New(t)

	inputCtx := testutil.NewTestContext()

	dynamo := &testutil.MockDynamoClient{}
	tableName := "watchers"

	sessionID := "abcdefg"

	facilitatorSessionKey := "bobsuruncle"

	facilitator := User{
		UserID:      "someUUIDFacilitator",
		Name:        "Bob",
		Handle:      "TheTester",
		CurrentVote: nil,
		SocketID:    "facilitator",
	}

	userA := User{
		UserID:      "someUUIDA",
		Name:        "A",
		CurrentVote: aws.String("1"),
		SocketID:    "aaaaaaaa",
	}

	userB := User{
		UserID:      "someUUIDB",
		Name:        "B",
		Handle:      "BBB",
		CurrentVote: aws.String("2"),
		SocketID:    "bbbbbbbb",
	}

	goneConnectionID := "gone"

	input := CompleteSessionView{
		SessionID:             sessionID,
		VotesShown:            true,
		FacilitatorSessionKey: facilitatorSessionKey,
		Facilitator:           facilitator,
		FacilitatorPoints:     true,
		Participants:          []User{userA, userB},
	}

	dispatchedMessages := make(map[string]api.Message, 0)
	dispatcher := api.MessageDispatcher(func(ctx context.Context, connectionID string, message api.Message) error {
		asserter.Equal(inputCtx, ctx)

		if connectionID == goneConnectionID {
			return errors.New("this errors out and we shouldn't choke on it")
		}

		if _, hasKey := dispatchedMessages[connectionID]; hasKey {
			asserter.Fail("duplicate dispatch")
		} else {
			dispatchedMessages[connectionID] = message
		}
		return nil
	})

	dynamo.On("QueryWithContext", inputCtx, &dynamodb.QueryInput{
		TableName: aws.String(tableName),
		KeyConditions: map[string]*dynamodb.Condition{
			"SessionID": {
				ComparisonOperator: aws.String("EQ"),
				AttributeValueList: []*dynamodb.AttributeValue{
					{S: aws.String(sessionID)},
				},
			},
		},
	}, emptyOpts).Return(&dynamodb.QueryOutput{
		Items: []map[string]*dynamodb.AttributeValue{
			{
				"SessionID": {S: aws.String("I'm the session record and don't have a socket")},
			},
			{
				"SocketID": {S: aws.String(facilitator.SocketID)},
			},
			{
				"SocketID": {S: aws.String(userA.SocketID)},
			},
			{
				"SocketID": {S: aws.String(userB.SocketID)},
			},
			{
				"SocketID": {S: aws.String(goneConnectionID)},
			},
		},
	}, nil)

	err := NewChangeNotifier(dynamo, tableName, dispatcher)(inputCtx, input)
	asserter.NoError(err)
	asserter.Equal(map[string]api.Message{
		userA.SocketID: {
			Type: "SESSION_UPDATED",
			Body: ParticipantSessionView{
				SessionID:  sessionID,
				VotesShown: true,
				Facilitator: User{
					UserID: facilitator.UserID,
					Handle: facilitator.Handle,
				},
				FacilitatorPoints: true,
				Participants: []User{
					{
						UserID:      userA.UserID,
						Name:        userA.Name,
						CurrentVote: userA.CurrentVote,
					},
					{
						UserID:      userB.UserID,
						Handle:      userB.Handle,
						CurrentVote: userB.CurrentVote,
					},
				},
			},
		},
		userB.SocketID: {
			Type: "SESSION_UPDATED",
			Body: ParticipantSessionView{
				SessionID:  sessionID,
				VotesShown: true,
				Facilitator: User{
					UserID: facilitator.UserID,
					Handle: facilitator.Handle,
				},
				FacilitatorPoints: true,
				Participants: []User{
					{
						UserID:      userA.UserID,
						Name:        userA.Name,
						CurrentVote: userA.CurrentVote,
					},
					{
						UserID:      userB.UserID,
						Handle:      userB.Handle,
						CurrentVote: userB.CurrentVote,
					},
				},
			},
		},
		facilitator.SocketID: {
			Type: "SESSION_UPDATED",
			Body: CompleteSessionView{
				SessionID:             sessionID,
				FacilitatorSessionKey: facilitatorSessionKey,
				VotesShown:            true,
				Facilitator: User{
					UserID:   facilitator.UserID,
					Name:     facilitator.Name,
					Handle:   facilitator.Handle,
					SocketID: facilitator.SocketID,
				},
				FacilitatorPoints: true,
				Participants: []User{
					{
						UserID:      userA.UserID,
						Name:        userA.Name,
						Handle:      userA.Handle,
						CurrentVote: userA.CurrentVote,
						SocketID:    userA.SocketID,
					},
					{
						UserID:      userB.UserID,
						Name:        userB.Name,
						Handle:      userB.Handle,
						CurrentVote: userB.CurrentVote,
						SocketID:    userB.SocketID,
					},
				},
			},
		},
	}, dispatchedMessages)
}

func Test_NewChangeNotifier_VotesHidden(t *testing.T) {
	asserter := assert.New(t)

	inputCtx := testutil.NewTestContext()

	dynamo := &testutil.MockDynamoClient{}
	tableName := "watchers"

	sessionID := "abcdefg"

	facilitatorSessionKey := "bobsuruncle"

	facilitator := User{
		UserID:      "someUUIDFacilitator",
		Name:        "Bob",
		Handle:      "TheTester",
		CurrentVote: nil,
		SocketID:    "facilitator",
	}

	userA := User{
		UserID:      "someUUIDA",
		Name:        "A",
		Handle:      "AAA",
		CurrentVote: aws.String("1"),
		SocketID:    "aaaaaaaa",
	}

	userB := User{
		UserID:      "someUUIDB",
		Name:        "B",
		Handle:      "BBB",
		CurrentVote: aws.String("2"),
		SocketID:    "bbbbbbbb",
	}

	goneConnectionID := "gone"

	input := CompleteSessionView{
		SessionID:             sessionID,
		VotesShown:            false,
		FacilitatorSessionKey: facilitatorSessionKey,
		Facilitator:           facilitator,
		FacilitatorPoints:     true,
		Participants:          []User{userA, userB},
	}

	dispatchedMessages := make(map[string]api.Message, 0)
	dispatcher := api.MessageDispatcher(func(ctx context.Context, connectionID string, message api.Message) error {
		asserter.Equal(inputCtx, ctx)

		if connectionID == goneConnectionID {
			return errors.New("this errors out and we shouldn't choke on it")
		}

		if _, hasKey := dispatchedMessages[connectionID]; hasKey {
			asserter.Fail("duplicate dispatch")
		} else {
			dispatchedMessages[connectionID] = message
		}
		return nil
	})

	dynamo.On("QueryWithContext", inputCtx, &dynamodb.QueryInput{
		TableName: aws.String(tableName),
		KeyConditions: map[string]*dynamodb.Condition{
			"SessionID": {
				ComparisonOperator: aws.String("EQ"),
				AttributeValueList: []*dynamodb.AttributeValue{
					{S: aws.String(sessionID)},
				},
			},
		},
	}, emptyOpts).Return(&dynamodb.QueryOutput{
		Items: []map[string]*dynamodb.AttributeValue{
			{
				"SessionID": {S: aws.String("I'm the session record and don't have a socket")},
			},
			{
				"SocketID": {S: aws.String(facilitator.SocketID)},
			},
			{
				"SocketID": {S: aws.String(userA.SocketID)},
			},
			{
				"SocketID": {S: aws.String(userB.SocketID)},
			},
			{
				"SocketID": {S: aws.String(goneConnectionID)},
			},
		},
	}, nil)

	err := NewChangeNotifier(dynamo, tableName, dispatcher)(inputCtx, input)
	asserter.NoError(err)
	asserter.Equal(map[string]api.Message{
		userA.SocketID: {
			Type: "SESSION_UPDATED",
			Body: ParticipantSessionView{
				SessionID:  sessionID,
				VotesShown: false,
				Facilitator: User{
					UserID: facilitator.UserID,
					Handle: facilitator.Handle,
				},
				FacilitatorPoints: true,
				Participants: []User{
					{
						UserID:      userA.UserID,
						Handle:      userA.Handle,
						CurrentVote: userA.CurrentVote,
					},
					{
						UserID: userB.UserID,
						Handle: userB.Handle,
					},
				},
			},
		},
		userB.SocketID: {
			Type: "SESSION_UPDATED",
			Body: ParticipantSessionView{
				SessionID:  sessionID,
				VotesShown: false,
				Facilitator: User{
					UserID: facilitator.UserID,
					Handle: facilitator.Handle,
				},
				FacilitatorPoints: true,
				Participants: []User{
					{
						UserID: userA.UserID,
						Handle: userA.Handle,
					},
					{
						UserID:      userB.UserID,
						Handle:      userB.Handle,
						CurrentVote: userB.CurrentVote,
					},
				},
			},
		},
		facilitator.SocketID: {
			Type: "SESSION_UPDATED",
			Body: CompleteSessionView{
				SessionID:             sessionID,
				FacilitatorSessionKey: facilitatorSessionKey,
				VotesShown:            false,
				Facilitator: User{
					UserID:   facilitator.UserID,
					Name:     facilitator.Name,
					Handle:   facilitator.Handle,
					SocketID: facilitator.SocketID,
				},
				FacilitatorPoints: true,
				Participants: []User{
					{
						UserID:      userA.UserID,
						Name:        userA.Name,
						Handle:      userA.Handle,
						CurrentVote: userA.CurrentVote,
						SocketID:    userA.SocketID,
					},
					{
						UserID:      userB.UserID,
						Name:        userB.Name,
						Handle:      userB.Handle,
						CurrentVote: userB.CurrentVote,
						SocketID:    userB.SocketID,
					},
				},
			},
		},
	}, dispatchedMessages)
}
