package models

import "github.com/graph-gophers/graphql-go"

type ComparisonOperator struct {
	Gt *string
	Ge *string
	Lt *string
	Le *string
	Eq *string
}

type ListSandboxesFilterInput struct {
	State         *[]*string
	AssignedUntil *ComparisonOperator
	AssignedSince *ComparisonOperator
}

type ListSandboxesFilter struct {
	State         []*string
	AssignedUntil ComparisonOperator
	AssignedSince ComparisonOperator
}

type SandboxInput struct {
	Id            graphql.ID
	AssignedTo    *string
	AssignedSince *string
	AssignedUntil *string
	Cloud         string
	State         *string
}
