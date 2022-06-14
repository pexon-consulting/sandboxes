package models

import "github.com/graph-gophers/graphql-go"

type AWS string
type AZURE string
type GCP string

type Cloud struct {
	AWS   AWS
	AZURE AZURE
	GCP   GCP
}

var PublicCloud = Cloud{
	AWS:   "AWS",
	AZURE: "AZURE",
	GCP:   "GCP",
}

func (c Cloud) GetAWS() string {
	return string(c.AWS)
}
func (c Cloud) GetAZURE() string {
	return string(c.AZURE)
}

type Sandbox struct {
	Result interface{}
}

// AwsSandbox and LeaseAwsResolver
type AwsSandbox struct {
	Id            graphql.ID
	AssignedTo    string
	AssignedSince string
	AssignedUntil string
	Cloud         string
	State         string
	Aws           AwsItem
}

type AwsResolver struct {
	U AwsSandbox
}

func (r *AwsResolver) Id() graphql.ID {
	return r.U.Id
}

func (r *AwsResolver) AssignedUntil() string {
	return r.U.AssignedUntil
}

func (r *AwsResolver) AssignedSince() string {
	return r.U.AssignedSince
}

func (r *AwsResolver) AssignedTo() string {
	return r.U.AssignedTo
}

func (r *AwsResolver) State() string {
	return r.U.State
}

// AzureSandbox and AzureResolver

type AzureSandbox struct {
	Id            graphql.ID
	PipelineId    string
	AssignedUntil string
	AssignedSince string
	AssignedTo    string
	Cloud         string
	Status        string
	ProjectId     string
	WebUrl        string
	SandboxName   string
	State         string
}

type AzureResolver struct {
	U AzureSandbox
}

func (r *AzureResolver) Id() graphql.ID {
	return r.U.Id
}

func (r *AzureResolver) PipelineId() string {
	return r.U.PipelineId
}

func (r *AzureResolver) AssignedUntil() string {
	return r.U.AssignedUntil
}

func (r *AzureResolver) AssignedSince() string {
	return r.U.AssignedSince
}

func (r *AzureResolver) AssignedTo() string {
	return r.U.AssignedTo
}
func (r *AzureResolver) SandboxName() string {
	return r.U.SandboxName
}

func (r *AzureResolver) State() string {
	return r.U.State
}

// ToAwsSandbox and ToAzureSandbox
func (r *Sandbox) ToAwsSandbox() (*AwsResolver, bool) {
	c, ok := r.Result.(*AwsResolver)
	return c, ok
}

func (r *Sandbox) ToAzureSandbox() (*AzureResolver, bool) {
	c, ok := r.Result.(*AzureResolver)
	return c, ok
}

// LIST SANDBOX

type ListSandboxes struct {
	U ListSandboxesResolver
}

type ListSandboxesResolver struct {
	Sandboxes []*Sandbox
}

func (r *ListSandboxes) Sandboxes() *[]*Sandbox {
	return &r.U.Sandboxes
}
