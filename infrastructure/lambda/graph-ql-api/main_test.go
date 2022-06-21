package main

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-lambda-go/events"
)

const jwtString string = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyLCJlbWFpbCI6Im1heEBtYXguZGUifQ.G9RcncmreaHa7XPoJZexiKHUBoBeVnjw3x3v9z4LPIU"

func TestHandler(t *testing.T) {
	ctx := context.TODO()
	request := events.APIGatewayProxyRequest{
		Headers: map[string]string{
			"Authorization: Bearer": jwtString,
		},
		Body: "",
	}

	result, err := Handler(ctx, request)

	resultErr := err == nil
	resultBody := result.Body == `{"errors":[{"message":"no operations in query document"}]}`
	statusCode := result.StatusCode == 200

	if !(statusCode && resultBody && resultErr) {
		t.Fatalf(fmt.Sprintf("expect statuscode 200 but got %v", result.StatusCode))
	}
}
