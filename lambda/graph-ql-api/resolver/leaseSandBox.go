package resolver

import (
	"context"
	"fmt"
	"lambda/aws-sandbox/graph-ql-api/api"
	"lambda/aws-sandbox/graph-ql-api/connection"
	"lambda/aws-sandbox/graph-ql-api/models"
	"lambda/aws-sandbox/graph-ql-api/utils"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/graph-gophers/graphql-go"
)

var valid bool

func (*Resolver) LeaseSandBox(ctx context.Context, args struct {
	LeaseTime string
	Cloud     string
}) (*models.Sandbox, error) {

	jwt, err = utils.RetrievJWTFromContext(ctx)
	if err != nil {
		// 🤦‍♀️
		return nil, fmt.Errorf("no valid jwt")

	}

	valid = utils.Lease_time_Input(args.LeaseTime)
	if !valid {
		// 🤦‍♀️
		return nil, fmt.Errorf("Lease-Time is not correct")
	}

	// generate a UUID
	id := uuid.New()
	graphqlId := graphql.ID(id.String())

	s := strings.Split(args.LeaseTime, "-")
	year, _ := strconv.Atoi(s[0])
	month, _ := strconv.Atoi(s[1])
	day, _ := strconv.Atoi(s[2])

	/*
		create current time and the until time object
	*/
	since, until := utils.TimeRange(year, month, day)

	svc := connection.GetEventBridgeClient(ctx)

	event := api.Event{
		Id:             string(graphqlId),
		Assigned_until: *until,
		Assigned_since: *since,
		User:           strings.ToLower(jwt.Payload.Email),
		Action:         "add",
		Cloud:          strings.ToLower(args.Cloud),
	}

	_, err := api.PutEvent(ctx, svc, &event)
	if err != nil {
		return nil, err
	}

	return &models.Sandbox{
		Result: &models.AwsResolver{
			U: models.AwsSandbox{
				Id:            graphqlId,
				AssignedUntil: *until,
				AssignedSince: *since,
				AssignedTo:    jwt.Payload.Email,
				State:         "requested",
			},
		},
	}, nil
}
