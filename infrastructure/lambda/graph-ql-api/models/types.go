package models

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
