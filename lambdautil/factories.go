package lambdautil

import (
	"os"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	awssession "github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/apigatewaymanagementapi"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-xray-sdk-go/xray"

	"github.com/jonsabados/pointypoints/api"
)

const SessionTimeout = time.Hour * 4

var SessionTable = os.Getenv("SESSION_TABLE")
var SessionSocketIndex = os.Getenv("SESSION_SOCKET_INDEX")

func CoreStartup() {
	err := xray.Configure(xray.Config{
		LogLevel: "warn",
	})
	if err != nil {
		panic(err)
	}
}

func NewProdMessageDispatcher() api.MessageDispatcher {
	gatewaysession, err := awssession.NewSession(&aws.Config{
		Region:   aws.String(os.Getenv("REGION")),
		Endpoint: aws.String(os.Getenv("GATEWAY_ENDPOINT")),
	})
	if err != nil {
		panic(err)
	}
	gateway := apigatewaymanagementapi.New(gatewaysession)
	xray.AWS(gateway.Client)
	return api.NewMessageDispatcher(gateway)
}

func DefaultAWSConfig() *awssession.Session {
	sess, err := awssession.NewSession(&aws.Config{})
	if err != nil {
		panic(err)
	}
	return sess
}

func NewDynamoClient(sess *awssession.Session) *dynamodb.DynamoDB {
	dynamo := dynamodb.New(sess)
	xray.AWS(dynamo.Client)
	return dynamo
}
