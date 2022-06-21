package relay

import (
	"context"
	"encoding/json"

	"github.com/aws/aws-lambda-go/events"
	"github.com/graph-gophers/graphql-go"
)

var global_ctx context.Context

type Handler struct {
	GraphqlSchema *graphql.Schema
}

type Params struct {
	Query         string                 `json:"query"`
	OperationName string                 `json:"operationName"`
	Variables     map[string]interface{} `json:"variables"`
}

type Middleware struct {
	Ctx     context.Context
	Schema  *graphql.Schema
	Params  Params
	Headers map[string]string
}

type MiddlewareResponse struct {
	Ctx      context.Context
	Pass     bool
	Response *events.APIGatewayProxyResponse
}

func (h *Handler) ServeHTTP(ctx context.Context, r events.APIGatewayProxyRequest, optFns ...func(Middleware) MiddlewareResponse) events.APIGatewayProxyResponse {
	params := Params{}
	json.Unmarshal([]byte(r.Body), &params)
	global_ctx = ctx

	for _, item := range optFns {
		middleware := Middleware{
			Ctx:     global_ctx,
			Schema:  h.GraphqlSchema,
			Params:  params,
			Headers: r.Headers,
		}
		x := item(middleware)

		if !x.Pass {
			return *x.Response
		}

		global_ctx = x.Ctx
	}

	response := h.GraphqlSchema.Exec(global_ctx, params.Query, params.OperationName, params.Variables)
	responseJSON, err := json.Marshal(response)

	if err != nil {
		return events.APIGatewayProxyResponse{
			Body:       "Internal Server Error",
			StatusCode: 500,
			Headers:    map[string]string{},
		}
	}

	headers := map[string]string{
		"Access-Control-Allow-Origin": "*",
		"Content-Type":                "application/json",
	}

	return events.APIGatewayProxyResponse{
		Body:       string(responseJSON),
		StatusCode: 200,
		Headers:    headers,
	}
}
