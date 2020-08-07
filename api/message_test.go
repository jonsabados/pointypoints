package api_test

import (
	"errors"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/apigatewaymanagementapi"
	"github.com/jonsabados/pointypoints/api"
	"github.com/jonsabados/pointypoints/session"
	"github.com/jonsabados/pointypoints/session/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"io/ioutil"
	"testing"
)

var emptyOpts []request.Option

func Test_NewMessageDispatcher_ErrorPosting(t *testing.T) {
	asserter := assert.New(t)

	inputCtx := testutil.NewTestContext()
	inputConnectionID := "weeee"
	inputMessage := api.Message{
		Type: "foo",
		Body: "bar",
	}

	expectedMessageContent, err := ioutil.ReadFile("fixture/errorMessage.json")
	if !asserter.NoError(err) {
		return
	}

	expectedError := "stuff went wrong"

	poster := &MockConnectionPoster{}
	poster.On("PostToConnectionWithContext", inputCtx, &apigatewaymanagementapi.PostToConnectionInput{
		ConnectionId: aws.String(inputConnectionID),
		Data:         expectedMessageContent,
	}, emptyOpts).Return(nil, errors.New(expectedError))

	err = api.NewMessageDispatcher(poster)(inputCtx, inputConnectionID, inputMessage)
	asserter.EqualError(err, expectedError)
}

func Test_NewMessageDispatcher(t *testing.T) {
	testCases := []struct {
		name                   string
		input                  api.Message
		expectedMessageFixture string
	}{
		{
			"session event sends as expected and does not include socket ids",
			api.Message{
				Type: "SESSION_UPDATED",
				Body: session.CompleteSessionView{
					SessionID:             "123",
					VotesShown:            true,
					FacilitatorSessionKey: "123345",
					Facilitator: session.User{
						UserID:      "a",
						Name:        "b",
						Handle:      "c",
						CurrentVote: aws.String("123"),
						SocketID:    "123",
					},
					FacilitatorPoints: false,
					Participants: []session.User{
						{
							UserID:      "f",
							Name:        "g",
							Handle:      "h",
							CurrentVote: aws.String("521"),
							SocketID:    "987",
						},
					},
				},
			},
			"fixture/sessionUpdate.json",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			asserter := assert.New(t)

			inputCtx := testutil.NewTestContext()
			inputConnectionID := "weeee"

			expectedMessageContent, err := ioutil.ReadFile(tc.expectedMessageFixture)
			if !asserter.NoError(err) {
				return
			}

			expectedError := "stuff went wrong"

			poster := &MockConnectionPoster{}
			poster.On("PostToConnectionWithContext", inputCtx, &apigatewaymanagementapi.PostToConnectionInput{
				ConnectionId: aws.String(inputConnectionID),
				Data:         expectedMessageContent,
			}, emptyOpts).Return(nil, errors.New(expectedError))

			err = api.NewMessageDispatcher(poster)(inputCtx, inputConnectionID, tc.input)
			asserter.EqualError(err, expectedError)
		})
	}
}

type MockConnectionPoster struct {
	mock.Mock
}

func (m *MockConnectionPoster) PostToConnectionWithContext(ctx aws.Context, input *apigatewaymanagementapi.PostToConnectionInput, opts ...request.Option) (*apigatewaymanagementapi.PostToConnectionOutput, error) {
	args := m.Called(ctx, input, opts)
	ret := args.Get(0)
	if ret == nil {
		return nil, args.Error(1)
	}
	return ret.(*apigatewaymanagementapi.PostToConnectionOutput), args.Error(1)
}
