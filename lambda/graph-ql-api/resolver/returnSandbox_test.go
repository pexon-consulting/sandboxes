package resolver_test

import (
	"context"
	"fmt"
	"lambda/aws-sandbox/graph-ql-api/api"
	"lambda/aws-sandbox/graph-ql-api/utils"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/eventbridge"
	eventbridgeTypes "github.com/aws/aws-sdk-go-v2/service/eventbridge/types"
	"github.com/graph-gophers/graphql-go/errors"
	"github.com/graph-gophers/graphql-go/gqltesting"
)

const ReturnSandboxQuery = `
			mutation ReturnSandbox($uuid: ID!) {
				returnSandbox(uuid: $uuid)
			}
	`

func TestReturnSandbox_AWS_Successfully_Requested(t *testing.T) {
	os.Setenv("env", "test")

	svc := api.MockedEventbridge{
		PutEvents_response: &eventbridge.PutEventsOutput{
			Entries:          []eventbridgeTypes.PutEventsResultEntry{},
			FailedEntryCount: 0,
		},
		PutEvents_err: nil,
	}

	svc_fail := api.MockedEventbridge{
		PutEvents_response: &eventbridge.PutEventsOutput{
			Entries:          []eventbridgeTypes.PutEventsResultEntry{},
			FailedEntryCount: 1,
		},
		PutEvents_err: nil,
	}

	gqltesting.RunTests(t, []*gqltesting.Test{
		{
			Context: context.WithValue(context.TODO(), utils.SvcClient, svc),
			Schema:  rootSchema,
			Variables: map[string]interface{}{
				"uuid": "123123",
			},
			Query:          ReturnSandboxQuery,
			ExpectedResult: `{"returnSandbox":true}`,
		},
		{
			Context: context.WithValue(context.TODO(), utils.SvcClient, svc_fail),
			Schema:  rootSchema,
			Variables: map[string]interface{}{
				"uuid": "123123",
			},
			Query:          ReturnSandboxQuery,
			ExpectedResult: `{"returnSandbox":null}`,
			ExpectedErrors: []*errors.QueryError{{
				ResolverError: fmt.Errorf(`there are failed events`),
				Message:       `there are failed events`,
				Path:          []interface{}{"returnSandbox"}}},
		},
	})
}
