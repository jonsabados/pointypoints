package api

import (
	"context"
	"encoding/json"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambdacontext"
	"github.com/google/uuid"
	"net/http"
)

type FieldValidationError struct {
	Field string `json:"field"`
	Error string `json:"error"`
}

type ValidationError struct {
	FieldErrors []FieldValidationError `json:"fieldErrors"`
	Errors      []string               `json:"errors"`
}

type Response struct {
	Result        interface{} `json:"result"`
	RequestID     string      `json:"requestId"`
}

func NewSuccessResponse(ctx context.Context, result interface{}) events.APIGatewayProxyResponse {
	return wrapResponse(Response{
		Result:        result,
		RequestID:     requestID(ctx),
	}, http.StatusOK)
}

func NewInternalServerError(ctx context.Context, ) events.APIGatewayProxyResponse {
	return wrapResponse(Response{
		Result:      "an internal server error has occurred",
		RequestID:   requestID(ctx),
	}, http.StatusInternalServerError)
}

func NewValidationFailureResponse(ctx context.Context, result ValidationError) events.APIGatewayProxyResponse {
	return wrapResponse(Response{
		Result:        result,
		RequestID: requestID(ctx),
	}, http.StatusBadRequest)
}

func NewPermissionDeniedResponse(ctx context.Context) events.APIGatewayProxyResponse {
	return wrapResponse(Response{
		Result:        "permission denied",
		RequestID: requestID(ctx),
	}, http.StatusForbidden)
}

func requestID(ctx context.Context) string {
	if awsCtx, inLambda := lambdacontext.FromContext(ctx); inLambda {
		return awsCtx.AwsRequestID
	} else {
		return uuid.New().String()
	}
}

func wrapResponse(r Response, statusCode int) events.APIGatewayProxyResponse {
	body, err := json.Marshal(r)
	if err != nil {
		panic(err) // if we can't marshal our own stuff then were hosed
	}
	return events.APIGatewayProxyResponse{
		StatusCode: statusCode,
		Body: string(body),
	}
}
