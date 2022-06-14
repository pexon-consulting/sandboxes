package models

type AwsItem struct {
	Account_id   string
	Account_name string
}
type AzureItem struct{}

type SandboxItem struct {
	Id             string     `dynamodbav:"id"`
	Assigned_to    string     `dynamodbav:"assigned_to"`
	Assigned_since string     `dynamodbav:"assigned_since"`
	Assigned_until string     `dynamodbav:"assigned_until"`
	Cloud          string     `dynamodbav:"cloud"`
	State          string     `dynamodbav:"state"`
	Aws            *AwsItem   `dynamodbav:"aws"`
	Azure          *AzureItem `dynamodbav:"azure"`
}
