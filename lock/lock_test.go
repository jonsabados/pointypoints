package lock

import (
	"context"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"sync"
	"testing"
	"time"
)

func Test_NewGlobalLockAppropriator_Locking(t *testing.T) {
	asserter := assert.New(t)

	tableName := uuid.New().String()

	cfg := aws.Config{
		Endpoint: aws.String("http://localhost:8000"),
		Region:   aws.String("us-east-1"),
	}
	sess := session.Must(session.NewSession())
	dynamo := dynamodb.New(sess, &cfg)
 	err := createLockTable(dynamo, tableName)
	if !asserter.NoError(err) {
		return
	}

	lockID := uuid.New().String()
	activeThreadMutex := sync.Mutex{}
	activeThreads := 0

	barrier := sync.WaitGroup{}
	barrier.Add(1)

	threadCount := 100

	done := sync.WaitGroup{}
	done.Add(threadCount)

	executionTimeout := time.Millisecond*time.Duration(threadCount) * 10000

	testInstance := NewGlobalLockAppropriator(dynamo, tableName, time.Nanosecond * 10, executionTimeout)

	for i := 0; i < threadCount; i++ {
		go func() {
			defer done.Done()
			barrier.Wait()

			ctx := context.Background()
			ctx, closeCtx := context.WithTimeout(ctx, executionTimeout)
			defer closeCtx()
			lock, err := testInstance(ctx, lockID)
			if !asserter.NoError(err) {
				return
			}

			activeThreadMutex.Lock()
			activeThreads++
			activeThreadMutex.Unlock()

			// without a sleep the test always passes even with no-op locking code
			time.Sleep(time.Millisecond * 2)

			activeThreadMutex.Lock()
			asserter.Equal(1, activeThreads)
			activeThreadMutex.Unlock()

			activeThreadMutex.Lock()
			activeThreads--
			activeThreadMutex.Unlock()

			err = lock.Unlock(ctx)
			asserter.NoError(err)
		}()
	}

	barrier.Done()
	done.Wait()
	asserter.Equal(0, activeThreads)

	_, err = dynamo.DeleteTable(&dynamodb.DeleteTableInput{
		TableName: aws.String(tableName),
	})
	asserter.NoError(err)
}

func Test_NewGlobalLockAppropriator_ContextTimeoutDuringLock(t *testing.T) {
	asserter := assert.New(t)

	tableName := uuid.New().String()

	cfg := aws.Config{
		Endpoint: aws.String("http://localhost:8000"),
		Region:   aws.String("us-east-1"),
	}
	sess := session.Must(session.NewSession())
	dynamo := dynamodb.New(sess, &cfg)
	err := createLockTable(dynamo, tableName)
	if !asserter.NoError(err) {
		return
	}

	lockID := uuid.New().String()

	testInstance := NewGlobalLockAppropriator(dynamo, tableName, time.Nanosecond * 10, time.Second)
	_, err = testInstance(context.Background(), lockID)
	if !asserter.NoError(err) {
		return
	}

	ctx := context.Background()
	ctx, ctxCancel := context.WithTimeout(ctx, time.Millisecond * 3)
	defer ctxCancel()

	_, err = testInstance(ctx, lockID)
	asserter.EqualError(err, "context closed")
	// at this point we should still have a lock record out there that's going to expire in ~ one second but will be left hanging by the privates lock, lets make sure we can lock at this point

	newCtx, newCtxCancel := context.WithTimeout(context.Background(), time.Second * 2)
	defer newCtxCancel()
	lock, err := testInstance(newCtx, lockID)
	asserter.NoError(err)
	asserter.NoError(lock.Unlock(newCtx))

	_, err = dynamo.DeleteTable(&dynamodb.DeleteTableInput{
		TableName: aws.String(tableName),
	})
	asserter.NoError(err)
}

func createLockTable(dynamo *dynamodb.DynamoDB, tableName string) error {
	_, err := dynamo.CreateTable(&dynamodb.CreateTableInput{
		TableName: aws.String(tableName),
		KeySchema: []*dynamodb.KeySchemaElement{
			{
				AttributeName: aws.String("LockID"),
				KeyType:       aws.String("HASH"),
			},
		},
		BillingMode: aws.String("PAY_PER_REQUEST"),
		AttributeDefinitions: []*dynamodb.AttributeDefinition{
			{
				AttributeName: aws.String("LockID"),
				AttributeType: aws.String("S"),
			},
		},
	})
	return err
}