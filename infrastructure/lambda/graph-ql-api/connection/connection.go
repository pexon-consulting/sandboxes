package connection

import (
	"context"
	"lambda/aws-sandbox/graph-ql-api/api"
	"lambda/aws-sandbox/graph-ql-api/utils"
	"log"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/eventbridge"
)

func GetDynamoDbClient(ctx context.Context) api.DynamoAPI {
	if os.Getenv("env") == "test" {
		svcUntyped := ctx.Value(utils.SvcClient)
		svc, b := svcUntyped.(api.MockedDynamoDB)
		if b {
			return svc
		} else {
			return api.MockedDynamoDB{}
		}

	}
	region := config.WithRegion("eu-central-1")
	optFns := []func(*config.LoadOptions) error{region}

	if os.Getenv("LOCAL_DYNAMODB") != "" {
		localhost := config.WithEndpointResolverWithOptions(aws.EndpointResolverWithOptionsFunc(
			func(service, region string, options ...interface{}) (aws.Endpoint, error) {
				return aws.Endpoint{URL: "http://localhost:8000", PartitionID: "aws", SigningRegion: "eu-central-1",
					HostnameImmutable: true}, nil
			}))
		optFns = append(optFns, localhost)
	}

	cfg, err := config.LoadDefaultConfig(ctx, optFns...)

	if err != nil {
		log.Fatalf("unable to load SDK config, %v", err)
	}

	// Using the Config value, create the DynamoDB client
	svc := dynamodb.NewFromConfig(cfg)

	return svc
}

func GetEventBridgeClient(ctx context.Context) api.EventbridgeAPI {
	if os.Getenv("env") == "test" {
		svcUntyped := ctx.Value(utils.SvcClient)
		svc, b := svcUntyped.(api.MockedEventbridge)
		if b {
			return svc
		} else {
			return api.MockedEventbridge{}
		}
	}

	sharedProfile := os.Getenv("sharedProfile")
	region := config.WithRegion("eu-central-1")

	optFns := []func(*config.LoadOptions) error{region}

	if sharedProfile != "" {
		profile := config.WithSharedConfigProfile(sharedProfile)
		optFns = append(optFns, profile)
	}

	cfg, err := config.LoadDefaultConfig(ctx, optFns...)

	if err != nil {
		log.Fatalf("unable to load SDK config, %v", err)
	}

	return eventbridge.NewFromConfig(cfg)

}
