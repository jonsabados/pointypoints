package api

import (
	"context"
	"github.com/aws/aws-lambda-go/lambdacontext"
	"github.com/google/uuid"
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
	RequestFailed bool        `json:"requestFailed"`
	FailureCode   string      `json:"failureCode,omitempty"`
	RequestID     string      `json:"requestId"`
}

func NewSuccessResponse(ctx context.Context, result interface{}) Response {
	return Response{
		Result:        result,
		RequestFailed: false,
		RequestID:     requestID(ctx),
	}
}

func NewInternalServerError(ctx context.Context, ) Response {
	return Response{
		Result:      "an internal server error has occurred",
		FailureCode: "error",
		RequestID:   requestID(ctx),
	}
}

func NewValidationFailureResponse(ctx context.Context, result ValidationError) Response {
	return Response{
		Result:        result,
		RequestFailed: true,
		FailureCode:   "validation",
		RequestID: requestID(ctx),
	}
}

func requestID(ctx context.Context) string {
	if awsCtx, inLambda := lambdacontext.FromContext(ctx); inLambda {
		return awsCtx.AwsRequestID
	} else {
		return uuid.New().String()
	}
}
