package api

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/eventbridge"
	"github.com/aws/aws-sdk-go-v2/service/eventbridge/types"
)

type Event struct {
	Id             string `json:"id"`
	User           string `json:"user"`
	Action         string `json:"action"`
	Assigned_until string `json:"assigned_until"`
	Assigned_since string `json:"assigned_since"`
	Cloud          string `json:"cloud"`
}

func PutEvent(ctx context.Context, svc EventbridgeAPI, event *Event) (*eventbridge.PutEventsOutput, error) {

	source := os.Getenv("event_source")
	eventBusName := os.Getenv("event_bus_name")

	b, err := json.Marshal(*event)
	if err != nil {
		return nil, err
	}

	input := &eventbridge.PutEventsInput{
		Entries: []types.PutEventsRequestEntry{
			{
				Source:       aws.String(source),
				DetailType:   aws.String("add event"),
				Detail:       aws.String(string(b)),
				EventBusName: aws.String(eventBusName),
			},
		},
	}

	result, err := svc.PutEvents(ctx, input)

	log.Println(&result.ResultMetadata)

	if err != nil {
		return nil, err
	}

	if result.FailedEntryCount != 0 {
		err := errors.New("there are failed events")
		return nil, err
	}

	return result, err

}
