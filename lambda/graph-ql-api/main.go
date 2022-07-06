package main

import (
	"context"
	"flag"
	"io/ioutil"
	"lambda/aws-sandbox/graph-ql-api/log"
	logger "lambda/aws-sandbox/graph-ql-api/log"
	"lambda/aws-sandbox/graph-ql-api/middleware"
	"lambda/aws-sandbox/graph-ql-api/relay"
	"lambda/aws-sandbox/graph-ql-api/resolver"
	"lambda/aws-sandbox/graph-ql-api/schema"
	"net/http"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/graph-gophers/graphql-go"
	"github.com/sirupsen/logrus"
)

func init() {
	logrus.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})
	logrus.SetLevel(logrus.DebugLevel)
}

func Handler(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {

	graphqlSchema := graphql.MustParseSchema(schema.GetRootSchema(), &resolver.Resolver{})

	relay := &relay.Handler{GraphqlSchema: graphqlSchema, Middleware: []relay.MiddlewareFunc{
		middleware.BindLogger,
		middleware.BindJwtToContext,
	}}

	logger := log.GetGlobalLogger(ctx).SetPackage("Main").Builder()

	logger.Info("test")

	return relay.ServeHTTP(ctx, request), nil
}

func local() {
	http.HandleFunc("/query", func(w http.ResponseWriter, r *http.Request) {

		b, err := ioutil.ReadAll(r.Body)

		if err != nil {
			panic(err)
		}

		ctx := context.TODO()
		mock_handler := events.APIGatewayProxyRequest{
			Headers: map[string]string{
				"Authorization":   strings.Join(r.Header["Authorization"], " "),
				"X-Amzn-Trace-Id": strings.Join(r.Header["X-Amzn-Trace-Id"], " "),
			},
			Body: string(b),
		}
		responseJSON, _ := Handler(ctx, mock_handler)

		w.Write([]byte(responseJSON.Body))
	})
	http.ListenAndServe(":8080", nil)
}

func main() {

	dev := flag.Bool("dev", false, "start local dev server")
	flag.Parse()
	if *dev {
		logConfig := logger.LogConfig{Package: "dev"}
		logConfig.SetPackage("dev").Builder().Info("start dev-server")
		local()
	}
	// request := events.APIGatewayProxyRequest{}
	// r, _ := Handler(context.TODO(), request)
	// fmt.Print(r)

	lambda.Start(Handler)
}
