package models_test

import (
	"fmt"
	"lambda/aws-sandbox/graph-ql-api/models"
	"testing"
)

func TestLeaseAwsResolver(t *testing.T) {
	tests := []models.AwsResolver{
		{
			models.AwsSandbox{
				Id: "Test",
			},
		},
		{
			models.AwsSandbox{
				Id:            "Test",
				AssignedSince: "Test",
				AssignedTo:    "Test",
				AssignedUntil: "Test",
				State:         "Test",
			},
		},
		{
			models.AwsSandbox{
				Id:            "",
				AssignedSince: "",
				AssignedTo:    "",
				AssignedUntil: "",
				State:         "",
			},
		},
	}

	for i, tt := range tests {
		t.Run(fmt.Sprintf("error in item %d expect", i+1), func(t *testing.T) {

			if tt.Id() != tt.U.Id {
				t.Errorf("error in item %d expect", i+1)
			}

			if tt.State() != tt.U.State {
				t.Errorf("error in item %d expect", i+1)
			}

			if tt.AssignedSince() != tt.U.AssignedSince {
				t.Errorf("error in item %d expect", i+1)
			}

			if tt.AssignedTo() != tt.U.AssignedTo {
				t.Errorf("error in item %d expect", i+1)
			}

			if tt.AssignedUntil() != tt.U.AssignedUntil {
				t.Errorf("error in item %d expect", i+1)
			}
		})
	}

}

func TestAzureResolver(t *testing.T) {
	tests := []models.AzureResolver{
		{
			models.AzureSandbox{
				Id:         "Test",
				PipelineId: "Test",
			},
		},
		{
			models.AzureSandbox{
				Id:            "Test",
				PipelineId:    "Test",
				AssignedSince: "Test",
				AssignedTo:    "Test",
				AssignedUntil: "Test",
			},
		},
		{
			models.AzureSandbox{
				Id:            "",
				PipelineId:    "",
				AssignedSince: "",
				AssignedTo:    "",
				AssignedUntil: "",
			},
		},
	}

	for i, tt := range tests {
		t.Run(fmt.Sprintf("error in item %d expect", i+1), func(t *testing.T) {

			if tt.Id() != tt.U.Id {
				t.Errorf("error in item %d expect", i+1)
			}

			if tt.PipelineId() != tt.U.PipelineId {
				t.Errorf("error in item %d expect", i+1)
			}

			if tt.AssignedSince() != tt.U.AssignedSince {
				t.Errorf("error in item %d expect", i+1)
			}

			if tt.AssignedTo() != tt.U.AssignedTo {
				t.Errorf("error in item %d expect", i+1)
			}

			if tt.AssignedUntil() != tt.U.AssignedUntil {
				t.Errorf("error in item %d expect", i+1)
			}
		})
	}

}

func TestCloudEnum(t *testing.T) {
	if models.PublicCloud.GetAWS() != "AWS" {
		t.Error("error in item dont match AWS")
	}
	if models.PublicCloud.GetAZURE() != "AZURE" {
		t.Error("error in item dont match AZURE")
	}
	if models.PublicCloud.AWS != "AWS" {
		t.Error("error in item dont match AWS")
	}
	if models.PublicCloud.AZURE != "AZURE" {
		t.Error("error in item dont match AZURE")
	}
}
