package resolver

import (
	"context"
	"fmt"
	"lambda/aws-sandbox/graph-ql-api/api"
	"lambda/aws-sandbox/graph-ql-api/connection"
	"lambda/aws-sandbox/graph-ql-api/log"
	"lambda/aws-sandbox/graph-ql-api/models"
	"lambda/aws-sandbox/graph-ql-api/utils"
)

func (*Resolver) UpdateSandbox(ctx context.Context, args struct {
	Sandbox models.SandboxInput
}) (*models.Sandbox, error) {
	logger := log.GetGlobalLogger(ctx).SetPackage("Resolver").SetResolver("UpdateSandbox").Builder()

	if args.Sandbox.AssignedUntil != nil && !utils.ValidateIsoTime(*args.Sandbox.AssignedUntil) {
		return nil, fmt.Errorf("assignedUntil validation Failed")
	}

	logger.Info("call UpdateSandbox")

	jwt, err = utils.RetrievJWTFromContext(ctx)

	if err != nil {
		return nil, err
	}

	svc := connection.GetEventBridgeClient(ctx)

	if args.Sandbox.AssignedUntil != nil {
		logger.Info("put UpdateSandbox event")
		event := api.Event{
			Action:         "update",
			User:           jwt.Payload.Email,
			Id:             string(args.Sandbox.Id),
			Assigned_until: string(*args.Sandbox.AssignedUntil),
		}
		_, err = api.PutEvent(ctx, svc, &event)
		if err != nil {
			return nil, err
		}
	}

	result := &models.Sandbox{}

	if args.Sandbox.Cloud == models.PublicCloud.GetAWS() {
		result.Result = &models.AwsResolver{
			U: models.AwsSandbox{
				Id:            args.Sandbox.Id,
				AssignedUntil: *args.Sandbox.AssignedUntil,
				AssignedTo:    jwt.Payload.Email,
			},
		}
	}

	if args.Sandbox.Cloud == models.PublicCloud.GetAZURE() {
		result.Result = &models.AzureResolver{
			U: models.AzureSandbox{
				Id:            args.Sandbox.Id,
				AssignedUntil: *args.Sandbox.AssignedUntil,
				AssignedTo:    jwt.Payload.Email,
			},
		}
	}

	return result, nil
}
