package middleware

import (
	"context"
	"lambda/aws-sandbox/graph-ql-api/log"
	"lambda/aws-sandbox/graph-ql-api/relay"
	"lambda/aws-sandbox/graph-ql-api/settings"
	"lambda/aws-sandbox/graph-ql-api/utils"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambdacontext"
)

func BindJwtToContext(i relay.Middleware) relay.MiddlewareResponse {
	logger := log.GetGlobalLogger(i.Ctx).SetPackage("Middleware").Builder()

	jwtString := i.Headers["Authorization"]
	jwtString = strings.Replace(jwtString, "Bearer ", "", 1)

	jwt, err := utils.ParseJWT(i.Ctx, jwtString)

	if jwtString == "" || err != nil {
		logger.Warn("no valid JWT provided")
		response := events.APIGatewayProxyResponse{
			Body:       "no JWT provided",
			StatusCode: 403,
			Headers:    map[string]string{},
		}
		return relay.MiddlewareResponse{Ctx: i.Ctx, Pass: false, Response: &response}
	}

	ctx := context.WithValue(i.Ctx, settings.Jwt, jwt)

	return relay.MiddlewareResponse{Ctx: ctx, Pass: true}
}

func BindLogger(i relay.Middleware) relay.MiddlewareResponse {

	logger := log.LogConfig{}

	if traceId := i.Headers["X-Amzn-Trace-Id"]; traceId != "" {
		logger = logger.SetField("TraceId", traceId)
	}

	if lc, lcbool := lambdacontext.FromContext(i.Ctx); lcbool {
		awsRequestID := lc.AwsRequestID
		logger = logger.SetField("AwsRequestId", awsRequestID)
	}

	ctx := context.WithValue(i.Ctx, settings.LogConfig, logger)

	return relay.MiddlewareResponse{Ctx: ctx, Pass: true}
}
