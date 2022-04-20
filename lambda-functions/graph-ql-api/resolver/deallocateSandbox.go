package resolver

import (
	"context"
	"lambda/aws-sandbox/graph-ql-api/models"
)

func (*Resolver) DeallocateSandbox(ctx context.Context, args struct {
	Account_id string
}) (*models.SandboxItem, error) {

	// svc := connection.GetDynamoDbClient(ctx)

	sandbox := models.SandboxItem{
		Account_id:     args.Account_id,
		Assigned_to:    "",
		Assigned_since: "",
		Assigned_until: "",
		Available:      "true",
	}
	return &sandbox, nil
	// updatedSandbox, err := api.UpdateSandBoxItem(ctx, svc, sandbox)

	// if err != nil {
	// 	return nil, err
	// }
	// return &models.SandBoxResolver{U: models.SandBoxResponse{
	// 	Message: "Sandbox successfully deallocate",
	// 	Sandbox: models.SandBoxItemResolver{
	// 		U: *updatedSandbox,
	// 	},
	// }}, nil
}
