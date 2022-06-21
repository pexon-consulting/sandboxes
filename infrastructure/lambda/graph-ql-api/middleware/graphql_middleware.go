package middleware

import (
	"context"
	"lambda/aws-sandbox/graph-ql-api/relay"
	"lambda/aws-sandbox/graph-ql-api/utils"
	"log"
	"strings"

	"github.com/aws/aws-lambda-go/events"
)

func BindJwtToContext(i relay.Middleware) relay.MiddlewareResponse {

	log.Println("HEADERS#################")
	log.Println(i.Headers)
	jwtString := i.Headers["Authorization"]
	jwtString = strings.Replace(jwtString, "Bearer ", "", 1)
	log.Println("jwt#################")
	log.Println(jwtString)
	jwt, err := utils.ParseJWT(jwtString)

	if jwtString == "" || err != nil {
		response := events.APIGatewayProxyResponse{
			Body:       "no JWT provided",
			StatusCode: 403,
			Headers:    map[string]string{},
		}
		return relay.MiddlewareResponse{Ctx: i.Ctx, Pass: false, Response: &response}
	}

	ctx := context.WithValue(i.Ctx, utils.Jwt, jwt)

	return relay.MiddlewareResponse{Ctx: ctx, Pass: true}
}
