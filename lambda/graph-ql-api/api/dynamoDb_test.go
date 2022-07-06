package api_test

import (
	"context"
	"fmt"
	"lambda/aws-sandbox/graph-ql-api/models"
	"testing"

	"time"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	dockerTypes "github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	dockerClient "github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"github.com/google/uuid"
)

// func TestQuerySandboxForUserErrorResponse(t *testing.T) {

// 	os.Setenv("dynamodb_table", "test")

// 	ctx := context.TODO()

// 	svc := api.MockedDynamoDB{
// 		Query_response: nil,
// 		Query_err:      fmt.Errorf("some Error"),
// 	}

// 	result := api.QuerySandboxForUser(ctx, svc, "", models.ListSandboxesFilter{})

// 	if len(result) != 0 {
// 		t.Fatalf("len not fit")
// 	}

// }

// func TestQuerySandboxForUserUnsetTable(t *testing.T) {

// 	os.Unsetenv("dynamodb_table")

// 	ctx := context.TODO()

// 	svc := api.MockedDynamoDB{
// 		Query_response: nil,
// 		Query_err:      fmt.Errorf("some Error"),
// 	}

// 	result := api.QuerySandboxForUser(ctx, svc, "", models.ListSandboxesFilter{})

// 	typeMatch := reflect.TypeOf(result) == reflect.TypeOf([]models.SandboxItem{})

// 	if typeMatch != true {
// 		k := reflect.TypeOf(result)
// 		p := reflect.TypeOf([]models.SandboxItem{})
// 		t.Fatalf("types not matching: %s, %s", k, p)
// 	}

// 	if len(result) != 0 {
// 		t.Fatalf("len not fit")
// 	}

// }

// func TestQuerySandboxForUserOneItem(t *testing.T) {

// 	os.Setenv("dynamodb_table", "test")

// 	ctx := context.TODO()

// 	items := []map[string]types.AttributeValue{
// 		{
// 			"id":             &types.AttributeValueMemberS{Value: "123"},
// 			"assigned_to":    &types.AttributeValueMemberS{Value: "123"},
// 			"assigned_since": &types.AttributeValueMemberS{Value: "123"},
// 			"assigned_until": &types.AttributeValueMemberS{Value: "123"},
// 			"state":          &types.AttributeValueMemberS{Value: "true"},
// 			"azure":          &types.AttributeValueMemberM{},
// 		},
// 		{
// 			"id":             &types.AttributeValueMemberS{Value: "123"},
// 			"assigned_to":    &types.AttributeValueMemberS{Value: "123"},
// 			"assigned_since": &types.AttributeValueMemberS{Value: "123"},
// 			"assigned_until": &types.AttributeValueMemberS{Value: "123"},
// 			"state":          &types.AttributeValueMemberS{Value: "true"},
// 			"aws": &types.AttributeValueMemberM{
// 				Value: map[string]types.AttributeValue{
// 					"account_id": &types.AttributeValueMemberS{Value: "my_id"},
// 				},
// 			},
// 		},
// 	}

// 	svc := api.MockedDynamoDB{
// 		Query_response: &dynamodb.QueryOutput{
// 			Count:            int32(len(items)),
// 			Items:            items,
// 			LastEvaluatedKey: map[string]types.AttributeValue{},

// 			ResultMetadata: middleware.Metadata{},
// 		},
// 		Query_err: nil,
// 	}

// 	result := api.QuerySandboxForUser(ctx, svc, "", models.ListSandboxesFilter{})

// 	typeMatch := reflect.TypeOf(result) == reflect.TypeOf([]models.SandboxItem{})

// 	if typeMatch != true {
// 		k := reflect.TypeOf(result)
// 		p := reflect.TypeOf([]models.SandboxItem{})
// 		t.Fatalf("types not matching: %s, %s", k, p)
// 	}

// 	if len(result) != len(items) {
// 		t.Fatalf("len should match, query result is %v but mock input is %v", len(result), len(items))
// 	}

// }

func MockDynamoDockerStart(ctx context.Context) (func(), func()) {

	cli, _ := dockerClient.NewClientWithOpts()

	imageName := "amazon/dynamodb-local:latest"

	config := container.Config{
		Image:        imageName,
		ExposedPorts: nat.PortSet{"8000": struct{}{}},
	}

	host := container.HostConfig{
		PortBindings: map[nat.Port][]nat.PortBinding{nat.Port("8000"): {{HostIP: "127.0.0.1", HostPort: "8000"}}},
	}

	resp, _ := cli.ContainerCreate(ctx, &config, &host, nil, nil, fmt.Sprintf("dynamoDB-%s", uuid.New()))

	ops := dockerTypes.ContainerStartOptions{}

	start := func() {
		cli.ContainerStart(ctx, resp.ID, ops)
	}
	stop := func() {
		var tim time.Duration = 20
		cli.ContainerStop(ctx, resp.ID, &tim)
	}

	return start, stop
}

func MockItems(input []models.SandboxItem) []types.WriteRequest {
	k := []types.WriteRequest{}

	for _, item := range input {
		p := types.WriteRequest{
			PutRequest: &types.PutRequest{
				Item: map[string]types.AttributeValue{
					"assigned_to":    &types.AttributeValueMemberS{Value: item.Assigned_to},
					"assigned_until": &types.AttributeValueMemberS{Value: item.Assigned_until},
					"assigned_since": &types.AttributeValueMemberS{Value: item.Assigned_since},
					"id":             &types.AttributeValueMemberS{Value: item.Id},
					"cloud":          &types.AttributeValueMemberS{Value: item.Cloud},
					"state":          &types.AttributeValueMemberS{Value: item.State},
				},
			},
		}
		k = append(k, p)
	}

	return k
}

/*
func TestMain(m *testing.M) {
	tableName := uuid.New().String()

	os.Setenv("DYNAMODB_TABLE", tableName)
	os.Setenv("LOCAL_DYNAMODB", "true")
	os.Setenv("AWS_ACCESS_KEY_ID", "dummy")
	os.Setenv("AWS_SECRET_ACCESS_KEY", "dummy")
	os.Setenv("AWS_SESSION_TOKEN", "dummy")

	ctx := context.TODO()
	start, stop := MockDynamoDockerStart(ctx)
	start()

	localhost := config.WithEndpointResolverWithOptions(aws.EndpointResolverWithOptionsFunc(
		func(service, region string, options ...interface{}) (aws.Endpoint, error) {
			return aws.Endpoint{URL: "http://localhost:8000", PartitionID: "aws", SigningRegion: "eu-central-1",
				HostnameImmutable: true}, nil
		}))

	cfg, _ := config.LoadDefaultConfig(ctx, config.WithRegion("eu-central-1"), localhost)

	svc := dynamodb.NewFromConfig(cfg)

	createTableInput := dynamodb.CreateTableInput{
		TableName:   aws.String(tableName),
		BillingMode: "PAY_PER_REQUEST",
		KeySchema: []types.KeySchemaElement{
			{
				AttributeName: aws.String("assigned_to"),
				KeyType:       "HASH",
			},
			{
				AttributeName: aws.String("id"),
				KeyType:       "SORT",
			},
		},
		AttributeDefinitions: []types.AttributeDefinition{
			{AttributeName: aws.String("assigned_to"),
				AttributeType: "S",
			},
			{AttributeName: aws.String("id"),
				AttributeType: "S",
			},
		},
	}

	svc.CreateTable(ctx, &createTableInput)

	batchWriteItemInput := dynamodb.BatchWriteItemInput{
		RequestItems: map[string][]types.WriteRequest{
			tableName: MockItems([]models.SandboxItem{
				{
					Id:             uuid.New().String(),
					Assigned_to:    "maximilian.haensel@pexon-consulting.de",
					Assigned_since: "2022-02-01T20:00:00",
					Assigned_until: "2022-03-01T20:00:00Z",
					Cloud:          "aws",
					State:          "accounted",
				},
				{
					Id:             uuid.New().String(),
					Assigned_to:    "maximilian.haensel@pexon-consulting.de",
					Assigned_since: "2022-03-01T20:00:00",
					Assigned_until: "2022-04-01T20:00:00Z",
					Cloud:          "aws",
					State:          "error",
				},
				{
					Id:             uuid.New().String(),
					Assigned_to:    "maximilian.haensel@pexon-consulting.de",
					Assigned_since: "2022-04-01T20:00:00",
					Assigned_until: "2022-05-01T20:00:00Z",
					Cloud:          "aws",
					State:          "error",
				},
			}),
		},
	}
	svc.BatchWriteItem(ctx, &batchWriteItemInput)

	code := m.Run()
	stop()
	defer stop()
	os.Exit(code)

}
*/
func stringPointer(s string) *string {
	return &s
}

func TestQuery(t *testing.T) {
	/*
		flag.Bool("tester", false, "")
		ctx := context.TODO()
		svc := connection.GetDynamoDbClient(ctx)

		user1 := "maximilian.haensel@pexon-consulting.de"
		_error := "error"
		_accounted := "accounted"

		tests := []struct {
			filter models.ListSandboxesFilter
			user   string
			result int
		}{
			{
				filter: models.ListSandboxesFilter{
					State: []*string{&_error},
				},
				user:   user1,
				result: 2,
			},
			{
				filter: models.ListSandboxesFilter{
					State: []*string{&_accounted},
				},
				user:   user1,
				result: 1,
			},
			{
				filter: models.ListSandboxesFilter{
					AssignedUntil: models.ComparisonOperator{
						Ge: stringPointer("2022-04-01T20:00:00Z"),
					},
				},
				user:   user1,
				result: 2,
			},
		}

		for i, tt := range tests {
			t.Run(fmt.Sprintf("testcase %d, query should return %d items", i, tt.result), func(t *testing.T) {

				result := api.QuerySandboxForUser(ctx, svc, tt.user, tt.filter)

				if len(result) != tt.result {
					t.Errorf(fmt.Sprintf("error in query %d is %d", len(result), tt.result))
				}
			})
		}
	*/
}
