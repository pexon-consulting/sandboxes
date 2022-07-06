package resolver

import (
	"context"
	"lambda/aws-sandbox/graph-ql-api/api"
	"lambda/aws-sandbox/graph-ql-api/connection"
	"lambda/aws-sandbox/graph-ql-api/utils"
)

var err error
var jwt utils.JwtItem

func (*Resolver) ReturnSandbox(ctx context.Context, args struct {
	Uuid string
}) (*bool, error) {

	jwt, err = utils.RetrievJWTFromContext(ctx)

	if err != nil {
		return nil, err
	}

	svc := connection.GetEventBridgeClient(ctx)

	event := api.Event{
		Action: "remove",
		User:   jwt.Payload.Email,
		Id:     args.Uuid,
	}

	_, err = api.PutEvent(ctx, svc, &event)

	if err != nil {
		return nil, err
	}

	result := true
	return &result, nil
}
