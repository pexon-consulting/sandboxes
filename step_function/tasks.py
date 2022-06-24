from aws_cdk import aws_stepfunctions as sfn, aws_stepfunctions_tasks as tasks, aws_dynamodb as dynamodb


def dynamo_db_query_task(app, id: str, table: dynamodb.Table):
    task = {
        "Type": "Task",
        "Parameters": {
            "TableName": table.table_name,
            "IndexName": "gsi_cloud",
            "Select": "ALL_PROJECTED_ATTRIBUTES",
            "KeyConditionExpression": "cloud = :cloud AND #order_state = :state",
            "FilterExpression": "aws.account_id = :account_id AND assigned_until > :assigned_until",
            "ExpressionAttributeValues": {
                ":cloud": {"S": "aws"},
                ":state": {"S": "accounted"},
                ":account_id": {"S.$": "$.Id"},
                ":assigned_until": {"S.$": "$$.Execution.StartTime"},
            },
            "ExpressionAttributeNames": {"#order_state": "state"},
        },
        "Resource": "arn:aws:states:::aws-sdk:dynamodb:query",
        "ResultPath": "$.input",
    }
    return sfn.CustomState(
        app,
        id,
        state_json=task,
    )


def organizations_list_accounts_for_parent(app, id: str, include_next_token: bool = False):
    task = {
        "Type": "Task",
        # TODO: remove hardcoded org-id
        "Parameters": {"ParentId": "ou-0vm6-7csr3fvu"},
        "Resource": "arn:aws:states:::aws-sdk:organizations:listAccountsForParent",
    }

    if include_next_token:
        task["Parameters"]["NextToken.$"] = "$.sso.NextToken"

    return sfn.CustomState(
        app,
        id,
        state_json=task,
    )


def sso_list_users_task(app, id: str):
    task = {
        "Type": "Task",
        "Parameters": {
            # TODO: hardcoded IdentityStoreId
            "IdentityStoreId": "d-99672b9ab3",
            "Filters": [{"AttributePath": "UserName", "AttributeValue.$": "$.event.user"}],
        },
        "Resource": "arn:aws:states:::aws-sdk:identitystore:listUsers",
        "OutputPath": "$.Users[0]",
    }

    return sfn.CustomState(
        app,
        id,
        state_json=task,
    )


def sso_create_assignment_task(app, id: str):
    task = {
        "Type": "Task",
        "Parameters": {
            # TODO: hardcoded vales
            "InstanceArn": "arn:aws:sso:::instance/ssoins-6987b5e3ca99dac9",
            "PermissionSetArn": "arn:aws:sso:::permissionSet/ssoins-6987b5e3ca99dac9/ps-f78cb46c16715cdf",
            "PrincipalId.$": "$.UserId",
            "PrincipalType": "USER",
            "TargetId.$": "$.sandbox.Id",
            "TargetType": "AWS_ACCOUNT",
        },
        "Resource": "arn:aws:states:::aws-sdk:ssoadmin:createAccountAssignment",
    }

    return sfn.CustomState(
        app,
        id,
        state_json=task,
    )


def sso_delete_account_assignment(app, id: str):
    task = {
        "Type": "Task",
        "Parameters": {
            "InstanceArn": "arn:aws:sso:::instance/ssoins-6987b5e3ca99dac9",
            "PermissionSetArn": "arn:aws:sso:::permissionSet/ssoins-6987b5e3ca99dac9/ps-f78cb46c16715cdf",
            "PrincipalId.$": "$.PrincipalId",
            "PrincipalType.$": "$.PrincipalType",
            "TargetId.$": "$.AccountId",
            "TargetType": "AWS_ACCOUNT",
        },
        "Resource": "arn:aws:states:::aws-sdk:ssoadmin:deleteAccountAssignment",
    }

    return sfn.CustomState(
        app,
        id,
        state_json=task,
    )


def sso_list_account_assignments(app, id: str):
    task = {
        "Type": "Task",
        "Parameters": {
            "AccountId.$": "$.aws.M.account_id.S",
            "InstanceArn": "arn:aws:sso:::instance/ssoins-6987b5e3ca99dac9",
            "PermissionSetArn": "arn:aws:sso:::permissionSet/ssoins-6987b5e3ca99dac9/ps-f78cb46c16715cdf",
        },
        "Resource": "arn:aws:states:::aws-sdk:ssoadmin:listAccountAssignments",
    }

    return sfn.CustomState(
        app,
        id,
        state_json=task,
    )


def dynamodb_query_expired_sandboxes(app, id: str, table: dynamodb.Table, cloud: str):
    task = {
        "Type": "Task",
        "Parameters": {
            "TableName": table.table_name,
            "IndexName": "gsi_cloud",
            "Select": "ALL_PROJECTED_ATTRIBUTES",
            "KeyConditionExpression": "cloud = :cloud AND #order_state = :state",
            "FilterExpression": "assigned_until <= :assigned_until",
            "ExpressionAttributeValues": {
                ":cloud": {"S": cloud},
                ":state": {"S": "accounted"},
                ":assigned_until": {"S.$": "$.time"},
            },
            "ExpressionAttributeNames": {"#order_state": "state"},
        },
        "Resource": "arn:aws:states:::aws-sdk:dynamodb:query",
        "ResultPath": "$.output",
    }
    return sfn.CustomState(
        app,
        id,
        state_json=task,
    )
