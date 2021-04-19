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
	Result    interface{} `json:"result"`
	RequestID string      `json:"requestId"`
}

func NewSuccessResponse(ctx context.Context, baseHeaders map[string]string, result interface{}) events.APIGatewayProxyResponse {
	return wrapResponse(Response{
		Result:    result,
		RequestID: requestID(ctx),
	}, responseHeaders(baseHeaders), http.StatusOK)
}

func NewInternalServerError(ctx context.Context, baseHeaders map[string]string) events.APIGatewayProxyResponse {
	return wrapResponse(Response{
		Result:    "an internal server error has occurred",
		RequestID: requestID(ctx),
	}, responseHeaders(baseHeaders), http.StatusInternalServerError)
}

func NewValidationFailureResponse(ctx context.Context, baseHeaders map[string]string, result ValidationError) events.APIGatewayProxyResponse {
	return wrapResponse(Response{
		Result:    result,
		RequestID: requestID(ctx),
	}, responseHeaders(baseHeaders), http.StatusBadRequest)
}

func NewPermissionDeniedResponse(ctx context.Context, baseHeaders map[string]string) events.APIGatewayProxyResponse {
	return wrapResponse(Response{
		Result:    "permission denied",
		RequestID: requestID(ctx),
	}, responseHeaders(baseHeaders), http.StatusForbidden)
}

func requestID(ctx context.Context) string {
	if awsCtx, inLambda := lambdacontext.FromContext(ctx); inLambda {
		return awsCtx.AwsRequestID
	} else {
		return uuid.New().String()
	}
}

func wrapResponse(r Response, headers map[string]string, statusCode int) events.APIGatewayProxyResponse {
	body, err := json.Marshal(r)
	if err != nil {
		panic(err) // if we can't marshal our own stuff then were hosed
	}
	return events.APIGatewayProxyResponse{
		StatusCode: statusCode,
		Body:       string(body),
		Headers:    headers,
	}
}

func responseHeaders(baseHeaders map[string]string) map[string]string {
	ret := make(map[string]string)
	for k, v := range baseHeaders {
		ret[k] = v
	}
	ret["content-type"] = "application/json"
	return ret
}
