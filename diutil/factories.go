package diutil

import (
	"github.com/aws/aws-sdk-go/aws"
	awssession "github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/apigatewaymanagementapi"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-xray-sdk-go/xray"
	"github.com/jonsabados/pointypoints/api"
	"github.com/jonsabados/pointypoints/lock"
	"os"
	"time"
)

func NewProdMessageDispatcher() api.MessageDispatcher {
	gatewaysession, err := awssession.NewSession(&aws.Config{
		Region:      aws.String(os.Getenv("REGION")),
		Endpoint:    aws.String(os.Getenv("GATEWAY_ENDPOINT")),
	})
	if err != nil {
		panic(err)
	}
	gateway := apigatewaymanagementapi.New(gatewaysession)
	xray.AWS(gateway.Client)
	return api.NewMessageDispatcher(gateway)
}

func NewGlobalLockAppropriator(dynamo *dynamodb.DynamoDB) lock.GlobalLockAppropriator {
	return lock.NewGlobalLockAppropriator(dynamo, os.Getenv("LOCK_TABLE"), time.Millisecond * 10, time.Second)
}