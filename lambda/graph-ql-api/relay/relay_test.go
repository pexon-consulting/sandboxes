package relay_test

import (
	"context"
	"fmt"
	"lambda/aws-sandbox/graph-ql-api/relay"
	"lambda/aws-sandbox/graph-ql-api/resolver"
	"lambda/aws-sandbox/graph-ql-api/schema"
	"testing"

	"github.com/aws/aws-lambda-go/events"
	"github.com/graph-gophers/graphql-go"
)

var request events.APIGatewayProxyRequest

type MockMiddlewareInput = struct {
	body       string
	statusCode int
	pass       bool
}

func MockMiddleware(mi MockMiddlewareInput) func(i relay.Middleware) relay.MiddlewareResponse {
	return func(i relay.Middleware) relay.MiddlewareResponse {
		response := events.APIGatewayProxyResponse{
			Body:       mi.body,
			StatusCode: mi.statusCode,
			Headers:    map[string]string{},
		}
		return relay.MiddlewareResponse{Ctx: i.Ctx, Pass: mi.pass, Response: &response}
	}
}

func TestServeMiddleware(t *testing.T) {

	graphqlSchema := graphql.MustParseSchema(schema.GetRootSchema(), &resolver.Resolver{})

	request = events.APIGatewayProxyRequest{
		Headers: map[string]string{
			"Authorization: Bearer": "",
		},
		Body: ``,
	}

	relay := &relay.Handler{GraphqlSchema: graphqlSchema}

	tests := []MockMiddlewareInput{
		{
			body: `{"errors":[{"message":"no operations in query document"}]}`, statusCode: 200, pass: true,
		},
		{
			body: "400 error", statusCode: 400, pass: false,
		},
		{
			body: "403 forbidden", statusCode: 403, pass: false,
		},
		{
			body: "500 server error", statusCode: 500, pass: false,
		},
	}

	for i, tt := range tests {
		t.Run(fmt.Sprintf("evaluate middleware-%d: %s ", i, tt.body), func(t *testing.T) {
			result := relay.ServeHTTP(context.TODO(), request, MockMiddleware(tt))
			resultBody := result.Body == tt.body
			statusCode := result.StatusCode == tt.statusCode
			if !(statusCode && resultBody) {
				t.Fatalf(fmt.Sprintf("expect statuscode 200 but got %v, expect resultBody %v", result.StatusCode, result.Body))
			}
		})
	}
}
