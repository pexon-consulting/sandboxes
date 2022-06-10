import json
from constructs import Construct


from variables import Enviroments

from aws_cdk import (
    Duration,
    aws_iam as iam,
    aws_stepfunctions as sfn,
    aws_stepfunctions_tasks as tasks,
    aws_dynamodb as dynamodb,
    aws_sqs as sqs,
)

from sso_step_function.tasks import (
    organizations_list_accounts_for_parent,
    dynamo_db_query_task,
    sso_list_users_task,
    sso_create_assignment_task,
    sso_delete_account_assignment,
    sso_list_account_assignments,
    dynamodb_query_expired_sandboxes,
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


class SsoStepFunctionAdd(Construct):
    def __init__(self, scope: Construct, construct_id: str, table: dynamodb.Table, enviroment, **kwargs) -> None:
        super().__init__(scope, construct_id, **kwargs)

        start_process = sfn.Parallel(
            self,
            "start process",
            result_selector={
                "sandboxes.$": "$.[1].sandboxes",
                "event": {"id.$": "$.[2].detail.id", "user.$": "$.[2].detail.user"},
            },
        )

        create_job_state = tasks.DynamoPutItem(
            self,
            "create job state",
            table=table,
            item={
                "assigned_to": tasks.DynamoAttributeValue.from_string(sfn.JsonPath.string_at("$.detail.user")),
                "id": tasks.DynamoAttributeValue.from_string(sfn.JsonPath.string_at("$.detail.id")),
                "cloud": tasks.DynamoAttributeValue.from_string("aws"),
                "state": tasks.DynamoAttributeValue.from_string("requested"),
                "assigned_until": tasks.DynamoAttributeValue.from_string(
                    sfn.JsonPath.string_at("$.detail.assigned_until")
                ),
                "assigned_since": tasks.DynamoAttributeValue.from_string(
                    sfn.JsonPath.string_at("$.detail.assigned_since")
                ),
            },
        )
        find_available_sandbox = FindAvailableSandbox(self, "find available sandbox", table=table)
        pass_data_start_process = sfn.Pass(self, "pass data start process")

        start_process.branch(create_job_state)
        start_process.branch(find_available_sandbox)
        start_process.branch(pass_data_start_process)

        create_sso_assignment = CreateSSOAssignment(self, "create sso assignment", table=table, enviroment=enviroment)

        chain = sfn.Chain.start(start_process).next(create_sso_assignment)

        sfn_ = sfn.StateMachine(self, "provison_sandbox", definition=chain, timeout=Duration.minutes(5))

        pol = iam.PolicyStatement(
            actions=[
                "organizations:ListAccountsForParent",
                "sso:CreateAccountAssignment",
                "identitystore:ListUsers",
                "sso:DeleteAccountAssignment",
            ],
            # TODO: add resources here
            resources=["*"],
        )

        dynamo_query = iam.PolicyStatement(
            actions=[
                "dynamodb:Query",
            ],
            resources=[table.table_arn, f"{table.table_arn}/*"],
        )

        sfn_.add_to_role_policy(pol)
        sfn_.add_to_role_policy(dynamo_query)
        self.step_function = sfn_


class SsoStepFunctionCronCleanUp(Construct):
    def __init__(
        self,
        scope: Construct,
        construct_id: str,
        table: dynamodb.Table,
        aws_nuke_queue: sqs.Queue,
        enviroment,
        **kwargs,
    ) -> None:
        super().__init__(scope, construct_id, **kwargs)

        dynamodb_query_expired_sandboxes_task = dynamodb_query_expired_sandboxes(
            self, "dynamodb_query_expired_sandboxes", table
        )

        map = sfn.Map(self, "map", items_path="$.output.Items", result_path=sfn.JsonPath.DISCARD)
        parallel = sfn.Parallel(self, "parallel")

        update_state = tasks.DynamoUpdateItem(
            self,
            "dynamoDB returned",
            table=table,
            expression_attribute_names={"#order_status": "state"},
            update_expression="SET #order_status = :state",
            expression_attribute_values={":state": tasks.DynamoAttributeValue.from_string("returned")},
            key={
                "assigned_to": tasks.DynamoAttributeValue.from_string(sfn.JsonPath.string_at("$.assigned_to.S")),
                "id": tasks.DynamoAttributeValue.from_string(sfn.JsonPath.string_at("$.id.S")),
            },
        )
        parallel.branch(update_state)

        send_message = tasks.SqsSendMessage(
            self,
            "nuke_aws",
            queue=aws_nuke_queue,
            message_body=sfn.TaskInput.from_object(
                {"cloud": "aws", "account_id": sfn.JsonPath.string_at("$.aws.M.account_id.S")}
            ),
        )
        parallel.branch(send_message)

        map2 = sfn.Map(self, "map2", items_path="$.AccountAssignments", result_path=sfn.JsonPath.DISCARD)
        sso_list_account_assignments_task = sso_list_account_assignments(self, "sso_list_account_assignments")

        sso_list_account_assignments_task.next(map2)
        sso_delete_account_assignment_task = sso_delete_account_assignment(self, "sso_delete_account_assignment")
        map2.iterator(sso_delete_account_assignment_task)

        parallel.branch(sso_list_account_assignments_task)

        map.iterator(parallel)
        chain = sfn.Chain.start(dynamodb_query_expired_sandboxes_task).next(map)

        sfn_ = sfn.StateMachine(self, "provison_sandbox", definition=chain, timeout=Duration.minutes(5))

        sso_policy = iam.PolicyStatement(
            actions=["sso:ListAccountAssignments", "sso:DeleteAccountAssignment"],
            # TODO: add resources here
            resources=["*"],
        )

        dynamo_query = iam.PolicyStatement(
            actions=[
                "dynamodb:Query",
            ],
            resources=[table.table_arn, f"{table.table_arn}/*"],
        )

        sfn_.add_to_role_policy(sso_policy)
        sfn_.add_to_role_policy(dynamo_query)
        self.step_function = sfn_
