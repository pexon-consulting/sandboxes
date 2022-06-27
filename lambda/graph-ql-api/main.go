package main

import (
	"context"
	"flag"
	"io/ioutil"
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
		middleware.BindJwtToContext,
		middleware.LogHeader,
	}}

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
				"Authorization": strings.Join(r.Header["Authorization"], " "),
			},
			Body: string(b),
		}
		responseJSON, _ := Handler(ctx, mock_handler)

		w.Write([]byte(responseJSON.Body))
	})
	http.ListenAndServe(":8080", nil)
}

func main() {
	log := logger.GetLogger(logger.LogConfig{Package: "main"})

	dev := flag.Bool("dev", false, "start local dev server")
	flag.Parse()
	if *dev {
		log.Info("start dev-server")
		local()
	}
	// request := events.APIGatewayProxyRequest{}
	// r, _ := Handler(context.TODO(), request)
	// fmt.Print(r)

	lambda.Start(Handler)
}
