package testutil

import (
	"context"
	"github.com/google/uuid"
)

func NewTestContext() context.Context {
	// just put something in it to make it unique for assertions
	return context.WithValue(context.Background(), "foo", uuid.New())
}