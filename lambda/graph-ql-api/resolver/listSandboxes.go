package resolver

import (
	"context"
	"fmt"
	"lambda/aws-sandbox/graph-ql-api/api"
	"lambda/aws-sandbox/graph-ql-api/connection"
	"lambda/aws-sandbox/graph-ql-api/log"
	"lambda/aws-sandbox/graph-ql-api/models"
	"lambda/aws-sandbox/graph-ql-api/utils"

	"github.com/graph-gophers/graphql-go"
)

func resolvFilter(f models.ListSandboxesFilter, inputF *models.ListSandboxesFilterInput) models.ListSandboxesFilter {
	if inputF == nil {
		return f
	}

	if *inputF.State != nil {
		f.State = *inputF.State
	}

	if inputF.AssignedUntil != nil {
		f.AssignedUntil = *inputF.AssignedUntil
	}

	if inputF.AssignedSince != nil {
		f.AssignedSince = *inputF.AssignedSince
	}
	return f

}

func (*Resolver) ListSandboxes(ctx context.Context, args struct {
	Filter *models.ListSandboxesFilterInput
}) (*models.ListSandboxes, error) {
	logger := log.GetGlobalLogger(ctx).SetPackage("Resolver").SetResolver("ListSandboxes").Builder()

	logger.Info("list Sandboxes")

	filter := resolvFilter(models.ListSandboxesFilter{}, args.Filter)

	jwt, err := utils.RetrievJWTFromContext(ctx)

	if err != nil {
		// 🤦‍♀️
		logger.Warn("cant retrive JWT from Context")
		return nil, fmt.Errorf("internal Server Error")
	}

	svc := connection.GetDynamoDbClient(ctx)

	items := api.QuerySandboxForUser(ctx, svc, jwt.Payload.Email, filter)

	sandboxes := []*models.Sandbox{}

	for _, item := range items {
		if item.Cloud == "aws" {
			toadd := &models.Sandbox{
				Result: &models.AwsResolver{
					U: models.AwsSandbox{
						Id:            graphql.ID(item.Id),
						AssignedUntil: item.Assigned_until,
						AssignedSince: item.Assigned_since,
						AssignedTo:    item.Assigned_to,
						State:         item.State,
						Cloud:         item.Cloud,
					},
				},
			}
			sandboxes = append(sandboxes, toadd)
		}
		if item.Cloud == "azure" {
			toadd := &models.Sandbox{
				Result: &models.AzureResolver{
					U: models.AzureSandbox{
						Id:            graphql.ID(item.Id),
						AssignedUntil: item.Assigned_until,
						AssignedSince: item.Assigned_since,
						AssignedTo:    item.Assigned_to,
						State:         item.State,
						Cloud:         item.Cloud,
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
