from constructs import Construct


from variables import Enviroments

from aws_cdk import (
    aws_stepfunctions as sfn,
    aws_stepfunctions_tasks as tasks,
    aws_dynamodb as dynamodb,
)

from step_function.tasks import (
    organizations_list_accounts_for_parent,
    dynamo_db_query_task,
    sso_list_users_task,
    sso_create_assignment_task,
)

sandbox_0_is_present = sfn.Condition.is_present("$.sandboxes[0]")


class FindAvailableSandbox(sfn.StateMachineFragment):
    def __init__(self, scope: Construct, construct_id: str, table: dynamodb.Table, **kwargs) -> None:
        super().__init__(scope, construct_id, **kwargs)

        org_list_accounts = organizations_list_accounts_for_parent(self, "list aws accounts")

        parallel = sfn.Parallel(self, "find sandbox", result_selector={"sso.$": "$.[0]", "sandboxes.$": "$.[1]"})

        org_list_accounts.next(parallel)

        map = sfn.Map(
            self,
            "map query sandboxes",
            max_concurrency=1,
            items_path=sfn.JsonPath.string_at("$.Accounts"),
            output_path="$..[?(@.input.Count==0)]",
        )

        query = dynamo_db_query_task(self, "query sandboxes", table=table)
        map.iterator(query)

        parallel.branch(sfn.Pass(self, "pass data find sandbox"))
        parallel.branch(map)

        choice = sfn.Choice(self, "if empty and next token")

        org_list_accounts_with_next_token = organizations_list_accounts_for_parent(
            self, "list aws accounts -recursiv", include_next_token=True
        )

        not_sandbox_item = sfn.Condition.not_(sandbox_0_is_present)
        next_token_present = sfn.Condition.is_present("$.sso.NextToken")

        choice.when(sfn.Condition.and_(not_sandbox_item, next_token_present), org_list_accounts_with_next_token)
        choice.otherwise(sfn.Succeed(self, "find sandbox finish"))

        org_list_accounts_with_next_token.next(parallel)
        parallel.next(choice)

        self._start_state = org_list_accounts
        self._end_states = choice.afterwards().end_states

    @property
    def start_state(self):
        return self._start_state

    @property
    def end_states(self):
        return self._end_states


class CreateSSOAssignment(sfn.StateMachineFragment):
    def __init__(self, scope: Construct, construct_id: str, table: dynamodb.Table, enviroment, **kwargs) -> None:
        super().__init__(scope, construct_id, **kwargs)

        if_sandbox_is_available = sfn.Choice(self, "if sandbox is available")

        if_sandbox_is_available.otherwise(sfn.Fail(self, "job failed"))

        dynamo_no_sandbox_available = tasks.DynamoUpdateItem(
            self,
            "dynamo no sandbox available",
            table=table,
            expression_attribute_names={"#order_status": "state"},
            update_expression="SET #order_status = :state",
            expression_attribute_values={":state": tasks.DynamoAttributeValue.from_string("no_sandbox_available")},
            key={
                "assigned_to": tasks.DynamoAttributeValue.from_string(sfn.JsonPath.string_at("$.event.user")),
                "id": tasks.DynamoAttributeValue.from_string(sfn.JsonPath.string_at("$.event.id")),
            },
        )

        if_sandbox_is_available.when(sfn.Condition.not_(sandbox_0_is_present), dynamo_no_sandbox_available)

        sso_list_users = sso_list_users_task(self, "sso list users")

        parallel_sso_list_users = sfn.Parallel(
            self,
            "parallel sso list users",
            result_selector={"sandbox.$": "$.[0].sandboxes[0]", "UserId.$": "$.[1].UserId", "event.$": "$[0].event"},
        )
        parallel_sso_list_users.branch(sfn.Pass(self, "pass data sso list users"))
        parallel_sso_list_users.branch(sso_list_users)

        sso_create_assignment_and_update_table = sfn.Parallel(self, "sso create assignment and update table")
        parallel_sso_list_users.next(sso_create_assignment_and_update_table)

        dynamo_update_status_and_account_id = tasks.DynamoUpdateItem(
            self,
            "dynamo update status and account id",
            table=table,
            key={
                "assigned_to": tasks.DynamoAttributeValue.from_string(sfn.JsonPath.string_at("$.event.user")),
                "id": tasks.DynamoAttributeValue.from_string(sfn.JsonPath.string_at("$.event.id")),
            },
            expression_attribute_names={"#order_status": "state"},
            update_expression="SET #order_status = :state, aws = :aws",
            expression_attribute_values={
                ":state": tasks.DynamoAttributeValue.from_string("accounted"),
                ":aws": tasks.DynamoAttributeValue.from_map(
                    {
                        "account_id": tasks.DynamoAttributeValue.from_string(sfn.JsonPath.string_at("$.sandbox.Id")),
                        "account_name": tasks.DynamoAttributeValue.from_string(
                            sfn.JsonPath.string_at("$.sandbox.Name")
                        ),
                    }
                ),
            },
        )

        sso_create_assignment = sso_create_assignment_task(self, "sso create assignment")

        sso_create_assignment_and_update_table.branch(dynamo_update_status_and_account_id)

        if enviroment == Enviroments.prod:
            sso_create_assignment_and_update_table.branch(sso_create_assignment)
        else:
            sso_create_assignment_and_update_table.branch(sfn.Pass(self, "dummy sso create assignment"))

        if_sandbox_is_available.when(sandbox_0_is_present, parallel_sso_list_users)

        self._start_state = if_sandbox_is_available
        self._end_states = if_sandbox_is_available.afterwards().end_states

    @property
    def start_state(self):
        return self._start_state

    @property
    def end_states(self):
        return self._end_states