package resolver

import (
	"context"
	"encoding/json"
	"fmt"
	"lambda/aws-sandbox/graph-ql-api/api"
	"lambda/aws-sandbox/graph-ql-api/connection"
	"lambda/aws-sandbox/graph-ql-api/models"
	"lambda/aws-sandbox/graph-ql-api/utils"
	"log"
	"net/http"
	"net/url"
	"os"
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
	logger.Info("call LeaseSandBox")

	jwt, err = utils.RetrievJWTFromContext(ctx)
	if err != nil {
		// 🤦‍♀️
		logger.Warn("no valid jwt")
		return nil, fmt.Errorf("no valid jwt")

	}

	valid = utils.Lease_time_Input(args.LeaseTime)
	if !valid {
		// 🤦‍♀️
		logger.Warn("Lease-Time is not correct")
		return nil, fmt.Errorf("Lease-Time is not correct")
	}

	// generate a UUID
	id := uuid.New()
	graphqlId := graphql.ID(id.String())

	s := strings.Split(args.LeaseTime, "-")
	year, _ := strconv.Atoi(s[0])
	month, _ := strconv.Atoi(s[1])
	day, _ := strconv.Atoi(s[2])

	// check if the Cloud is AZURE
	if args.Cloud == models.PublicCloud.GetAZURE() {
		// do your logic here 🤡
		since, until := utils.TimeRange(year, month, day)
		state_name := strings.Replace(strings.Split(jwt.Payload.Email, "@")[0], ".", "-", 1)
		sandbox_name := "rg-bootcamp-" + state_name
		data := url.Values{
			"rg_name":       {sandbox_name},
			"trainee_email": {jwt.Payload.Email},
			"removal_date":  {*until},
			"created_by":    {jwt.Payload.Email},
		}

		res := models.GitlabPipelineResponse{}
		url := os.Getenv("gitlab_azure_pipeline_webhook")
		url += "&variables[TF_STATE_NAME]=" + state_name

		resp, err := http.PostForm(url, data)
		if err != nil {
			log.Fatal(err)
			return nil, err
		}
		json.NewDecoder(resp.Body).Decode(&res)

		svc := connection.GetEventBridgeClient(ctx)

		event := api.Event{}

		_, err = api.PutEvent(ctx, svc, &event)

		if err != nil {
			return nil, err
		}

		return &models.Sandbox{
			Result: &models.AzureResolver{
				U: models.AzureSandbox{
					Id:            graphql.ID(uuid.New().String()),
					AssignedUntil: *until,
					AssignedSince: *since,
					AssignedTo:    jwt.Payload.Email,
					PipelineId:    strconv.Itoa(res.Id),
					Status:        res.Status,
					ProjectId:     strconv.Itoa(res.ProjectId),
					WebUrl:        res.WebUrl,
					SandboxName:   sandbox_name,
				},
			},
		}, nil
	}

	// check if the Cloud is AWS
	if args.Cloud == models.PublicCloud.GetAWS() {

		jwt, err = utils.RetrievJWTFromContext(ctx)

		/*
			create current time and the until time object
		*/
		since, until := utils.TimeRange(year, month, day)

		svc := connection.GetEventBridgeClient(ctx)

		event := api.Event{
			Id:             string(graphqlId),
			Assigned_until: *until,
			Assigned_since: *since,
			User:           jwt.Payload.Email,
			Action:         "add",
			Cloud:          "aws",
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

	// TODO 🔥 add custom error to be more clear whats going on
	return nil, fmt.Errorf("internal servererror")
}
