package api_test

import (
	"context"
	"errors"
	"fmt"
	"lambda/aws-sandbox/graph-ql-api/api"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/eventbridge"
	"github.com/aws/aws-sdk-go-v2/service/eventbridge/types"
	"github.com/google/uuid"
)

type mockedPutRequest struct {
	svc       api.MockedEventbridge
	errorType error
}

func TestPutEvent(t *testing.T) {
	os.Setenv("env", "test")
	// os.Setenv("sharedProfile", "sandbox")
	os.Setenv("event_source", "de.pexon.sso.app")
	os.Setenv("event_bus_name", "testSandboxEventBusbus4E22DCFC")

	tests := []mockedPutRequest{
		{
			api.MockedEventbridge{
				PutEvents_response: &eventbridge.PutEventsOutput{
					Entries:          []types.PutEventsResultEntry{},
					FailedEntryCount: 0,
				},
				PutEvents_err: nil,
			},
			nil,
		},
		{
			api.MockedEventbridge{
				PutEvents_response: &eventbridge.PutEventsOutput{
					Entries:          []types.PutEventsResultEntry{},
					FailedEntryCount: 1,
				},
				PutEvents_err: nil,
			},
			errors.New("there are failed events"),
		},
	}

	for i, tt := range tests {
		t.Run(fmt.Sprintf("testcase %d", i), func(t *testing.T) {

			ctx := context.TODO()

			event := api.Event{
				Id:             uuid.NewString(),
				User:           "maximilian.haensel@pexon-consulting.de",
				Action:         "add",
				Assigned_until: "time",
				Assigned_since: "time",
			}

			_, err := api.PutEvent(ctx, tt.svc, &event)

			if err != nil && err.Error() != tt.errorType.Error() {
				t.Error(tt.errorType.Error())
				t.Error(err.Error())
			}
		})
	}

}
