package resolver

import (
	"context"
	"lambda/aws-sandbox/graph-ql-api/api"
	"lambda/aws-sandbox/graph-ql-api/connection"
	"lambda/aws-sandbox/graph-ql-api/models"

	"github.com/graph-gophers/graphql-go"
)

func (*Resolver) ListSandboxes(ctx context.Context, args struct {
	Email string
}) (*models.ListSandboxes, error) {

	svc := connection.GetDynamoDbClient(ctx)

	items := api.ScanSandboxTable(ctx, svc)

	sandboxes := []*models.Sandbox{}

	aws := models.Cloud{}.GetAWS()
	azure := models.Cloud{}.GetAZURE()

	for _, item := range items {
		if item.Cloud == aws {
			toadd := &models.Sandbox{
				Result: &models.AwsResolver{
					U: models.AwsSandbox{
						Id:            graphql.ID(item.Id),
						AssignedUntil: item.Assigned_until,
						AssignedSince: item.Assigned_since,
						AssignedTo:    item.Assigned_to,
						State:         item.State,
						Cloud:         models.Cloud{}.GetAWS(),
					},
				},
			}
			sandboxes = append(sandboxes, toadd)
		}
		if item.Cloud == azure {
			toadd := &models.Sandbox{
				Result: &models.AzureResolver{
					U: models.AzureSandbox{
						Id:            graphql.ID(item.Id),
						AssignedUntil: item.Assigned_until,
						AssignedSince: item.Assigned_since,
						AssignedTo:    item.Assigned_to,
						State:         item.State,
						Cloud:         azure,
					},
				},
			}
			sandboxes = append(sandboxes, toadd)
		}
	}

	return &models.ListSandboxes{
		U: models.ListSandboxesResolver{
			Sandboxes: sandboxes,
		},
	}, nil
}
