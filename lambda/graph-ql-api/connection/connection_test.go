package connection_test

import (
	"context"
	"fmt"
	"lambda/aws-sandbox/graph-ql-api/api"
	"lambda/aws-sandbox/graph-ql-api/connection"
	"lambda/aws-sandbox/graph-ql-api/settings"
	"os"
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/eventbridge"
)

func TestGetDynamoDbClientLiveClient(t *testing.T) {

	tests := []struct {
		env    string
		result bool
	}{
		{
			env:    "prod",
			result: true,
		},
		{
			env:    "test",
			result: false,
		},
	}

	for i, tt := range tests {
		t.Run(fmt.Sprintf("testcase %d, contain %s items", i, tt.env), func(t *testing.T) {
			os.Setenv("env", tt.env)

			svc := connection.GetDynamoDbClient(context.TODO())
			result := reflect.TypeOf(svc) == reflect.TypeOf(&dynamodb.Client{})

			os.Unsetenv("env")
			if result != tt.result {
				t.Errorf(fmt.Sprintf("error in item %t is %t", result, tt.result))
			}
		})
	}
}

func TestGetDynamoDbClientMockClient(t *testing.T) {
	tests := []struct {
		env    string
		result bool
	}{
		{
			env:    "prod",
			result: false,
		},
		{
			env:    "test",
			result: true,
		},
	}

	for i, tt := range tests {
		t.Run(fmt.Sprintf("testcase %d, contain %s items", i, tt.env), func(t *testing.T) {
			os.Setenv("env", tt.env)

			ctx := context.WithValue(context.TODO(), settings.SvcClient, api.MockedDynamoDB{})
			svc := connection.GetDynamoDbClient(ctx)

			result := reflect.TypeOf(svc) == reflect.TypeOf(api.MockedDynamoDB{})

			os.Unsetenv("env")
			if result != tt.result {
				t.Errorf(fmt.Sprintf("error in item %t is %t", result, tt.result))
			}
		})
	}
}

func TestGetEventBridgeClientLiveClient(t *testing.T) {

	tests := []struct {
		env    string
		result bool
	}{
		{
			env:    "prod",
			result: true,
		},
		{
			env:    "test",
			result: false,
		},
	}

	for i, tt := range tests {
		t.Run(fmt.Sprintf("testcase %d, contain %s items", i, tt.env), func(t *testing.T) {
			os.Setenv("env", tt.env)

			svc := connection.GetEventBridgeClient(context.TODO())
			result := reflect.TypeOf(svc) == reflect.TypeOf(&eventbridge.Client{})

			os.Unsetenv("env")
			if result != tt.result {
				t.Errorf(fmt.Sprintf("error in item %t is %t", result, tt.result))
			}
		})
	}
}

func TestGetEventBridgeClientMockClient(t *testing.T) {
	tests := []struct {
		env    string
		result bool
	}{
		{
			env:    "prod",
			result: false,
		},
		{
			env:    "test",
			result: true,
		},
	}

	for i, tt := range tests {
		t.Run(fmt.Sprintf("testcase %d, contain %s items", i, tt.env), func(t *testing.T) {
			os.Setenv("env", tt.env)

			ctx := context.WithValue(context.TODO(), settings.SvcClient, api.MockedEventbridge{})
			svc := connection.GetEventBridgeClient(ctx)

			result := reflect.TypeOf(svc) == reflect.TypeOf(api.MockedEventbridge{})

			os.Unsetenv("env")
			if result != tt.result {
				t.Errorf(fmt.Sprintf("error in item %t is %t", result, tt.result))
			}
		})
	}
}
