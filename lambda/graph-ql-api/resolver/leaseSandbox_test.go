package resolver_test

import (
	"context"
	"fmt"
	"lambda/aws-sandbox/graph-ql-api/api"
	"lambda/aws-sandbox/graph-ql-api/resolver"
	"lambda/aws-sandbox/graph-ql-api/schema"
	"lambda/aws-sandbox/graph-ql-api/utils"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/aws/smithy-go/middleware"
	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/errors"
	"github.com/graph-gophers/graphql-go/gqltesting"
)

func init() {
	os.Setenv("env", "test")
	os.Setenv("DYNAMODB_TABLE", "test")
}

var rootSchema = graphql.MustParseSchema(schema.GetRootSchema(), &resolver.Resolver{})

var no_scan_result = api.MockedDynamoDB{
	Scan_response: &dynamodb.ScanOutput{
		Count:            0,
		Items:            []map[string]types.AttributeValue{},
		LastEvaluatedKey: map[string]types.AttributeValue{},
		ScannedCount:     0,
		ResultMetadata:   middleware.Metadata{},
	},
	Scan_err: nil,
}

var no_available_sandbox = api.MockedDynamoDB{
	Scan_response: &dynamodb.ScanOutput{
		Count: 2,
		Items: []map[string]types.AttributeValue{
			{
				"available": &types.AttributeValueMemberS{Value: "false"},
			},
			{
				"available": &types.AttributeValueMemberS{Value: "false"},
			},
		},
		LastEvaluatedKey: map[string]types.AttributeValue{},
		ScannedCount:     2,
		ResultMetadata:   middleware.Metadata{},
	},
	Scan_err: nil,
}

var malformed_available = api.MockedDynamoDB{
	Scan_response: &dynamodb.ScanOutput{
		Count: 2,
		Items: []map[string]types.AttributeValue{
			{
				"available": &types.AttributeValueMemberS{Value: "xxz"},
			},
		},
		LastEvaluatedKey: map[string]types.AttributeValue{},
		ScannedCount:     2,
		ResultMetadata:   middleware.Metadata{},
	},
	Scan_err: nil,
}

var Query = `
			mutation LeaseSandBox($leaseTime: String!,  $cloud: Cloud!) {
				leaseSandBox( leaseTime: $leaseTime, cloud: $cloud) {
					__typename
					... on CloudSandbox {
						assignedTo
						assignedUntil
						state
					}
					... on AwsSandbox {
					}
					... on AzureSandbox {
					}
				}
			}
	`

/*
	############################################################

	Test-Suits for all Clouds

	############################################################
*/

func TestLeaseASandbox_malformed_input(t *testing.T) {
	// error responses
	// path := []interface{}{"leaseSandBox"}
	// var noValideMail = []*errors.QueryError{{
	// 	ResolverError: fmt.Errorf(`no valid Pexon-Mail`),
	// 	Message:       `no valid Pexon-Mail`,
	// 	Path:          path}}

	// var wrongLeaseTime = []*errors.QueryError{{
	// 	ResolverError: fmt.Errorf(`Lease-Time is not correct`),
	// 	Message:       `Lease-Time is not correct`,
	// 	Path:          path}}

	tests := []struct {
		testname       string
		svc            api.MockedDynamoDB
		variables      map[string]interface{}
		ExpectedErrors []*errors.QueryError
	}{
		// {
		// 	testname: "wrongLeaseTime",
		// 	svc:      no_scan_result,
		// 	variables: map[string]interface{}{
		// 		"leaseTime": "2024",
		// 		"cloud":     "AWS",
		// 	},
		// 	// ExpectedErrors: wrongLeaseTime,
		// },
	}

	for i, test := range tests {
		t.Run(fmt.Sprintf("testcase %d, testname %s", i, test.testname), func(t *testing.T) {
			ctx := context.TODO()
			ctx = context.WithValue(ctx, utils.SvcClient, test.svc)
			gqltesting.RunTest(t, &gqltesting.Test{
				Context:        ctx,
				Schema:         rootSchema,
				Variables:      test.variables,
				Query:          Query,
				ExpectedResult: "null",
				ExpectedErrors: test.ExpectedErrors,
			})
		})
	}
}

func TestLeaseSandbox_Internal_Servererror(t *testing.T) {

	gqltesting.RunTests(t, []*gqltesting.Test{
		{
			Context: context.TODO(),
			Schema:  rootSchema,
			Variables: map[string]interface{}{
				"leaseTime": "2024-05-02",
				"cloud":     "GCP",
			},
			Query:          Query,
			ExpectedResult: "null",
			ExpectedErrors: []*errors.QueryError{{
				ResolverError: fmt.Errorf(`no valid jwt`),
				Message:       `no valid jwt`,
				Path:          []interface{}{"leaseSandBox"}}},
		},
	})
}

/*
	############################################################

	Test-Suits for AWS-Sandbox-calls

	############################################################
*/

/*
func TestLeaseSandbox_AWS_Successfully_Requested(t *testing.T) {
	os.Setenv("env", "test")

	svc := api.MockedEventbridge{
		PutEvents_response: &eventbridge.PutEventsOutput{
			Entries:          []eventbridgeTypes.PutEventsResultEntry{},
			FailedEntryCount: 0,
		},
		PutEvents_err: nil,
	}

	ctx := context.WithValue(context.TODO(), utils.SvcClient, svc)

	gqltesting.RunTests(t, []*gqltesting.Test{
		{
			Context: ctx,
			Schema:  rootSchema,
			Variables: map[string]interface{}{
				"email":     "test.test@pexon-consulting.de",
				"leaseTime": "2024-05-02",
				"cloud":     "AWS",
			},
			Query: Query,
			ExpectedResult: `{
				"leaseSandBox":{
					"__typename": "AwsSandbox",
					"assignedTo": "test.test@pexon-consulting.de",
					"assignedUntil": "2024-05-01T22:00:00Z",
					"state": "requested"
				}
			}`,
		},
	})
}
/*
func TestLeaseSandbox_AWS_Successfully_Fail_Requested(t *testing.T) {
	os.Setenv("env", "test")

	svc := api.MockedEventbridge{
		PutEvents_response: &eventbridge.PutEventsOutput{
			Entries:          []eventbridgeTypes.PutEventsResultEntry{},
			FailedEntryCount: 1,
		},
		PutEvents_err: nil,
	}

	ctx := context.WithValue(context.TODO(), utils.SvcClient, svc)

	gqltesting.RunTests(t, []*gqltesting.Test{
		{
			Context: ctx,
			Schema:  rootSchema,
			Variables: map[string]interface{}{
				"email":     "test.test@pexon-consulting.de",
				"leaseTime": "2024-05-02",
				"cloud":     "AWS",
			},
			Query:          Query,
			ExpectedResult: "null",
			ExpectedErrors: []*errors.QueryError{{
				ResolverError: fmt.Errorf(`there are failed events`),
				Message:       `there are failed events`,
				Path:          []interface{}{"leaseSandBox"}}},
		},
	})
}

/*
	############################################################

	Test-Suits for Azure-Sandbox-calls

	currently disabled because they will fail

	############################################################
*/

// func TestLeaseSandbox_AZURE_Successfully_Provisioning(t *testing.T) {

// 	gqltesting.RunTests(t, []*gqltesting.Test{
// 		{
// 			Context: context.TODO(),
// 			Schema:  rootSchema,
// 			Variables: map[string]interface{}{
// 				"email":     "test.test@pexon-consulting.de",
// 				"leaseTime": "2024-05-02",
// 				"cloud":     "AZURE",
// 			},
// 			Query: Query,
// 			ExpectedResult: `{
// 				"leaseSandBox":{
// 					"__typename": "AzureSandbox",
// 					"pipelineId": "this-is-azure",
// 					"assignedSince": "2022",
// 					"assignedTo": "max",
// 					"assignedUntil": "2023",
// 					"id": "this-azure2"
// 				}
// 			}`,
// 		},
// 	})
// }
