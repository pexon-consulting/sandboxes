package api

import (
	"fmt"
	"lambda/aws-sandbox/graph-ql-api/models"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

func TestComparisonOperatorFilterBuilder(t *testing.T) {
	input_value := "2011"
	attr := "testAttr"

	tests := []struct {
		filter   models.ComparisonOperator
		operator string
	}{
		{
			filter:   models.ComparisonOperator{Gt: &input_value},
			operator: ">",
		},
		{
			filter:   models.ComparisonOperator{Ge: &input_value},
			operator: ">=",
		},
		{
			filter:   models.ComparisonOperator{Eq: &input_value},
			operator: "=",
		},

		{
			filter:   models.ComparisonOperator{Le: &input_value},
			operator: "<=",
		},
		{
			filter:   models.ComparisonOperator{Lt: &input_value},
			operator: "<",
		},
		{
			filter:   models.ComparisonOperator{Gt: &input_value, Lt: &input_value},
			operator: "<",
		},
	}

	for _, tt := range tests {
		t.Run("testcase", func(t *testing.T) {
			expression, expressionKey, value := ComparisonOperatorFilterBuilder(tt.filter, attr)

			if expression == nil || expressionKey == nil || value == nil {
				t.Error("Values should not be nil")
				t.Fail()
			} else {
				if expectedExpression := fmt.Sprintf("%s %s :%s", attr, tt.operator, attr); *expression != expectedExpression {
					t.Errorf(fmt.Sprintf("expect Expression: %s but got %s", expectedExpression, *expression))
					t.Fail()
				}
				if expectedExpressionKey := fmt.Sprintf(":%s", attr); *expressionKey != expectedExpressionKey {
					t.Errorf(fmt.Sprintf("expect ExpressionKey: %s but got %s", expectedExpressionKey, *expressionKey))
					t.Fail()
				}
				if *value != input_value {
					t.Errorf("expect %s but got %s", input_value, *value)
					t.Fail()
				}
			}
		})
	}
}

func TestComparisonOperatorFilterBuilderWithEmptyFilter(t *testing.T) {
	expression, expressionKey, value := ComparisonOperatorFilterBuilder(models.ComparisonOperator{}, "")

	if expression != nil && expressionKey != nil && value != nil {
		t.Error("an empty filter should return nil")
	}

}

func TestC(t *testing.T) {

	filter1 := models.ListSandboxesFilter{}

	filter := comparisonOperatorFilterInput{
		filter:                    filter1,
		attributes:                map[string]string{"asd": "asd"},
		filterExpression:          []string{"asdas"},
		expressionAttributeValues: map[string]types.AttributeValue{"asda": &types.AttributeValueMemberS{Value: "asd"}},
	}

	expression, expressionKey := ComparisonOperatorFilter(filter)

	t.Log(expression)
	t.Log(expressionKey)
	t.Fail()
	// if expression != nil && expressionKey != nil && value != nil {
	// 	t.Error("an empty filter should return nil")
	// }

}
