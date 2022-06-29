package api

import (
	"context"
	"fmt"
	"lambda/aws-sandbox/graph-ql-api/log"
	"lambda/aws-sandbox/graph-ql-api/models"
	"os"
	"reflect"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

type comparisonOperatorFilterInput struct {
	filter                    models.ListSandboxesFilter
	attributes                map[string]string
	filterExpression          []string
	expressionAttributeValues map[string]types.AttributeValue
}

func ComparisonOperatorFilter(input comparisonOperatorFilterInput) ([]string, map[string]types.AttributeValue) {
	r := reflect.ValueOf(input.filter)
	for key, operator := range input.attributes {
		f := reflect.Indirect(r).FieldByName(operator)
		if f.IsValid() {
			comparisonFilter := f.Interface().(models.ComparisonOperator)
			expression, expressionKey, value := ComparisonOperatorFilterBuilder(comparisonFilter, key)
			if expression != nil {
				input.filterExpression = append(input.filterExpression, *expression)
				input.expressionAttributeValues[*expressionKey] = &types.AttributeValueMemberS{Value: *value}
			}
		}
	}

	return input.filterExpression, input.expressionAttributeValues
}

func ComparisonOperatorFilterBuilder(filter models.ComparisonOperator, attribute string) (*string, *string, *string) {
	var model models.ComparisonOperator
	if filter != model {
		for key, operator := range map[string]string{
			"Eq": "=",
			"Le": "<=",
			"Lt": "<",
			"Ge": ">=",
			"Gt": ">",
		} {
			r := reflect.ValueOf(filter)
			f := reflect.Indirect(r).FieldByName(key)
			value := f.Interface().(*string)
			if value != nil {
				expressionKey := fmt.Sprintf(":%s", attribute)
				expression := fmt.Sprintf("%s %s %s", attribute, operator, expressionKey)
				return &expression, &expressionKey, value
			}
		}
	}
	return nil, nil, nil
}

func QuerySandboxForUser(ctx context.Context, svc DynamoAPI, email string, filter models.ListSandboxesFilter) (items []models.SandboxItem) {
	logger := log.GetGlobalLogger(ctx).SetPackage("Api").SetFunction("QuerySandboxForUser").Builder()

	items = []models.SandboxItem{}

	defer func() {
		// defer function to escape the panic
		if r := recover(); r != nil {
			logger.Error(fmt.Sprintf("QuerySandboxForUser: %o", r))
			items = []models.SandboxItem{}
		}
	}()

	table := os.Getenv("DYNAMODB_TABLE")

	if len(table) == 0 {
		logger.Error("failed to find table-name, env-variable DYNAMODB_TABLE is empty")
		return items
	}

	ExpressionAttributeValues := map[string]types.AttributeValue{
		":assigned_to": &types.AttributeValueMemberS{Value: email},
	}
	var FilterExpression []string = []string{}
	var ExpressionAttributeNames map[string]string = make(map[string]string)
	var arrayExpression []string = []string{}

	if len(filter.State) != 0 {
		logger.Info("Build State Filter")
		for _, item := range filter.State {
			key := fmt.Sprintf(":key_%s", *item)
			ExpressionAttributeValues[key] = &types.AttributeValueMemberS{Value: strings.ToLower(*item)}
			arrayExpression = append(arrayExpression, key)
		}
		stateExpression := fmt.Sprintf("#order_state IN (%s)", strings.Join(arrayExpression, ", "))
		FilterExpression = append(FilterExpression, stateExpression)
		ExpressionAttributeNames["#order_state"] = "state"
	}

	FilterExpression, ExpressionAttributeValues = ComparisonOperatorFilter(comparisonOperatorFilterInput{
		filter: filter,
		attributes: map[string]string{
			"assigned_until": "AssignedUntil",
			"assigned_since": "AssignedSince",
		},
		filterExpression:          FilterExpression,
		expressionAttributeValues: ExpressionAttributeValues,
	})

	input := dynamodb.QueryInput{
		TableName:                 aws.String(table),
		Select:                    types.Select(types.SelectAllAttributes),
		KeyConditionExpression:    aws.String("assigned_to = :assigned_to"),
		ExpressionAttributeValues: ExpressionAttributeValues,
	}

	if len(FilterExpression) != 0 {
		logger.Info("Build FilterExpression")
		k := strings.Join(FilterExpression, " AND ")
		input.FilterExpression = aws.String(k)
	}

	if len(ExpressionAttributeNames) != 0 {
		logger.Info("attach ExpressionAttributeNames")
		input.ExpressionAttributeNames = ExpressionAttributeNames
	}

	logger.Info("run dynamoDB-Query")
	query, err := svc.Query(ctx, &input)

	if err != nil {
		logger.Error(fmt.Sprintf("fail to Query DynamoDB %s", err.Error()))
		return items
	}

	err = attributevalue.UnmarshalListOfMaps(query.Items, &items)

	if err != nil {
		logger.Error(fmt.Sprintf("fail to unmarshal Dynamodb Scan Items, %s", err.Error()))
		return []models.SandboxItem{}
	}
	return items
}

/*
func UpdateSandBoxItem(ctx context.Context, svc DynamoAPI, sandbox models.SandboxItem) (*models.SandboxItem, error) {

	table := os.Getenv("dynamodb_table")

	if sandbox.Id == "" {
		return nil, fmt.Errorf("no Account_id provided")
	}

	if len(table) == 0 {
		err := fmt.Errorf("env-variable dynamodb_table is empty")
		log.Print(fmt.Errorf("ERROR: failed to find table-name %v", err))
		return nil, err
	}

	key := map[string]types.AttributeValue{
		"account_id": &types.AttributeValueMemberS{Value: sandbox.Id},
	}

	update := struct {
		Assigned_to    string `dynamodbav:":assigned_to"`
		Assigned_since string `dynamodbav:":assigned_since"`
		Assigned_until string `dynamodbav:":assigned_until"`
		Available      string `dynamodbav:":available"`
	}{
		Assigned_to:    sandbox.Assigned_to,
		Assigned_since: sandbox.Assigned_since,
		Assigned_until: sandbox.Assigned_until,
		Available:      sandbox.State,
	}

	expr, err := attributevalue.MarshalMap(update)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal Record, %w", err)
	}

	updateExpression := aws.String(`
		SET
		assigned_to = :assigned_to,
		assigned_since = :assigned_since,
		assigned_until = :assigned_until,
		available = :available`,
	)

	input := &dynamodb.UpdateItemInput{
		UpdateExpression:          updateExpression,
		TableName:                 aws.String(table),
		ExpressionAttributeValues: expr,
		Key:                       key,
		ReturnValues:              "ALL_NEW",
	}

	response, err := svc.UpdateItem(ctx, input)

	if err != nil {
		return nil, err
	}

	p := models.SandboxItem{}
	attributevalue.UnmarshalMap(response.Attributes, &p)

	return &p, nil
}
*/

/*

func ScanSandboxTable(ctx context.Context, svc DynamoAPI) []models.SandboxItem {

	items := []models.SandboxItem{}

	table := os.Getenv("dynamodb_table")

	if len(table) == 0 {
		err := fmt.Errorf("env-variable dynamodb_table is empty")
		log.Print(fmt.Errorf("ERROR: failed to find table-name %v", err))
		return items
	}

	scanInput := dynamodb.ScanInput{
		TableName: aws.String(table),
	}

	scan, err := svc.Scan(ctx, &scanInput)

	if err != nil {
		log.Print(fmt.Errorf("ERROR: failed to Scan DynamoDB %v", err))
		return items
	}

	err = attributevalue.UnmarshalListOfMaps(scan.Items, &items)

	if err != nil {
		log.Print(fmt.Errorf("ERROR: failed to unmarshal Dynamodb Scan Items, %v", err))
		return items
	}
	return items
}

*/
