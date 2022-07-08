import json
from constructs import Construct


from variables import Enviroments

from aws_cdk import (
    Duration,
    RemovalPolicy,
    aws_iam as iam,
    aws_stepfunctions as sfn,
    aws_stepfunctions_tasks as tasks,
    aws_dynamodb as dynamodb,
    aws_sqs as sqs,
    aws_lambda as _lambda,
    aws_logs as logs,
)

from step_function.tasks import (
    sso_delete_account_assignment,
    sso_list_account_assignments,
    dynamodb_query_expired_sandboxes,
)

from step_function.aws_tasks import FindAvailableSandbox, CreateSSOAssignment

AWS_PREFIX = "aws"
AZURE_PREFIX = "azure"
sandbox_0_is_present = sfn.Condition.is_present("$.sandboxes[0]")


class StepFunctionTemplate(Construct):
    def __init__(self, scope: Construct, id: str, chain: sfn.Chain, **kwargs) -> None:
        super().__init__(scope, id, **kwargs)
        log_group = logs.LogGroup(
            self,
            f"Logs",
            retention=logs.RetentionDays.TWO_MONTHS,
            removal_policy=RemovalPolicy.DESTROY,
        )
        sfn_ = sfn.StateMachine(
            self,
            f"{id}",
            definition=chain,
            timeout=Duration.minutes(5),
            tracing_enabled=True,
            logs=sfn.LogOptions(destination=log_group, level=sfn.LogLevel.ERROR),
        )
        log_group
        self.sfn = sfn_


class DynamoTasks(Construct):
    def __init__(self, scope: Construct, construct_id: str, table: dynamodb.Table, enviroment=None, **kwargs) -> None:
        super().__init__(scope, construct_id, **kwargs)
        self.table = table

    def put_sandbox(self, cloud: str):
        return tasks.DynamoPutItem(
            self,
            "create job state",
            table=self.table,
            item={
                "assigned_to": tasks.DynamoAttributeValue.from_string(sfn.JsonPath.string_at("$.detail.user")),
                "id": tasks.DynamoAttributeValue.from_string(sfn.JsonPath.string_at("$.detail.id")),
                "cloud": tasks.DynamoAttributeValue.from_string(cloud),
                "state": tasks.DynamoAttributeValue.from_string("requested"),
                "assigned_until": tasks.DynamoAttributeValue.from_string(
                    sfn.JsonPath.string_at("$.detail.assigned_until")
                ),
                "assigned_since": tasks.DynamoAttributeValue.from_string(
                    sfn.JsonPath.string_at("$.detail.assigned_since")
                ),
            },
        )


class AzureStepFunctionAdd(Construct):
    def __init__(
        self,
        scope: Construct,
        construct_id: str,
        table: dynamodb.Table,
        function: _lambda.Function,
        enviroment,
        **kwargs,
    ) -> None:
        super().__init__(scope, construct_id, **kwargs)
        PREFIX = "azure"

        start_process = sfn.Parallel(self, "start process", result_path="$.parallel")

        create_job = DynamoTasks(scope, f"create_job", table, enviroment).put_sandbox(AZURE_PREFIX)
        start_process.branch(create_job)
        api_proxy = tasks.LambdaInvoke(self, "lambda-invoke", lambda_function=function, input_path="$.detail")
        start_process.branch(api_proxy)

        dynamo_update_azure_response = tasks.DynamoUpdateItem(
            self,
            "dynamo update status and account id",
            table=table,
            key={
                "assigned_to": tasks.DynamoAttributeValue.from_string(sfn.JsonPath.string_at("$.detail.user")),
                "id": tasks.DynamoAttributeValue.from_string(sfn.JsonPath.string_at("$.detail.id")),
            },
            expression_attribute_names={"#order_status": "state"},
            update_expression="SET #order_status = :state, azure = :azure",
            expression_attribute_values={
                ":state": tasks.DynamoAttributeValue.from_string("accounted"),
                ":azure": tasks.DynamoAttributeValue.map_from_json_path("$.parallel[1].Payload"),
                # tasks.DynamoAttributeValue.from_map(
                #     {
                #         "pipeline-id": tasks.DynamoAttributeValue.from_string(
                #             sfn.JsonPath.string_at("$.parallel[1].Payload.pipeline-id")
                #         ),
                #         "azure-magic": tasks.DynamoAttributeValue.from_string(
                #             sfn.JsonPath.string_at("$.parallel[1].Payload.azure-magic")
                #         ),
                #     }
                # ),
            },
        )

        chain = sfn.Chain.start(start_process).next(dynamo_update_azure_response)

        # log_group = logs.LogGroup(self, "LogGroupProvisonAzSandbox", retention=logs.RetentionDays.TWO_MONTHS)
        # sfn_ = sfn.StateMachine(
        #     self,
        #     "provison_az_sandbox",
        #     definition=chain,
        #     timeout=Duration.minutes(5),
        #     tracing_enabled=True,
        #     logs=sfn.LogOptions(destination=log_group, level=sfn.LogLevel.ERROR),
        # )

        self.step_function = StepFunctionTemplate(self, "provison_az_sandbox", chain=chain).sfn


class AwsStepFunctionAdd(Construct):
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

        # sfn_ = sfn.StateMachine(
        #     self, "provison_sandbox", definition=chain, timeout=Duration.minutes(5), tracing_enabled=True
        # )

        sfn_ = StepFunctionTemplate(self, "ProvisonAwsSandbox", chain=chain).sfn

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


class MultiCloudSfnUpdateSandbox(Construct):
    def __init__(self, scope: Construct, construct_id: str, table: dynamodb.Table, enviroment, **kwargs) -> None:
        super().__init__(scope, construct_id, **kwargs)
        PREFIX = "MultiCloud"

        update_assigned_until = tasks.DynamoUpdateItem(
            self,
            f"{PREFIX} dynamoDB update",
            table=table,
            update_expression="SET assigned_until = :assigned_until",
            expression_attribute_values={
                ":assigned_until": tasks.DynamoAttributeValue.from_string(
                    sfn.JsonPath.string_at("$.detail.assigned_until")
                )
            },
            key={
                "assigned_to": tasks.DynamoAttributeValue.from_string(sfn.JsonPath.string_at("$.detail.user")),
                "id": tasks.DynamoAttributeValue.from_string(sfn.JsonPath.string_at("$.detail.id")),
            },
        )

        chain = sfn.Chain.start(update_assigned_until)

        self.step_function = StepFunctionTemplate(self, "MultiCloudSfnUpdateSandbox", chain=chain).sfn


"""
Cleanup Sandboxes 
"""


class AWSRemoveWithID(sfn.StateMachineFragment):
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

        PREFIX = "aws"

        parallel = sfn.Parallel(self, f"{PREFIX} parallel")

        update_state = tasks.DynamoUpdateItem(
            self,
            f"{PREFIX} dynamoDB returned",
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
            f"{PREFIX} sqs nuke aws",
            queue=aws_nuke_queue,
            message_body=sfn.TaskInput.from_object(
                {"cloud": "aws", "account_id": sfn.JsonPath.string_at("$.aws.M.account_id.S")}
            ),
        )
        parallel.branch(send_message)

        map = sfn.Map(
            self,
            f"{PREFIX} map AccountAssignments",
            items_path="$.AccountAssignments",
            result_path=sfn.JsonPath.DISCARD,
        )

        sso_list_account_assignments_task = sso_list_account_assignments(
            self, f"{PREFIX} sso_list_account_assignments"
        )

        sso_list_account_assignments_task.next(map)
        sso_delete_account_assignment_task = sso_delete_account_assignment(
            self, f"{PREFIX} sso_delete_account_assignment"
        )
        map.iterator(sso_delete_account_assignment_task)

        if enviroment == Enviroments.prod:
            parallel.branch(sso_list_account_assignments_task)
        else:
            parallel.branch(sfn.Pass(self, "dummy account_assignments_task"))

        self._start_state = parallel
        self._end_states = parallel.end_states

    @property
    def start_state(self):
        return self._start_state

    @property
    def end_states(self):
        return self._end_states


class AzureRemoveWithID(sfn.StateMachineFragment):
    def __init__(
        self, scope: Construct, construct_id: str, table: dynamodb.Table, function: _lambda.Function, **kwargs
    ) -> None:
        super().__init__(scope, construct_id, **kwargs)

        PREFIX = "azure"

        api_proxy = tasks.LambdaInvoke(self, "lambda-invoke", lambda_function=function)

        update_state = tasks.DynamoUpdateItem(
            self,
            f"{PREFIX} dynamoDB returned",
            table=table,
            expression_attribute_names={"#order_status": "state"},
            update_expression="SET #order_status = :state",
            expression_attribute_values={":state": tasks.DynamoAttributeValue.from_string("returned")},
            key={
                "assigned_to": tasks.DynamoAttributeValue.from_string(sfn.JsonPath.string_at("$.assigned_to.S")),
                "id": tasks.DynamoAttributeValue.from_string(sfn.JsonPath.string_at("$.id.S")),
            },
        )
        api_proxy.next(update_state)

        self._start_state = api_proxy
        self._end_states = update_state.end_states

    @property
    def start_state(self):
        return self._start_state

    @property
    def end_states(self):
        return self._end_states


class MultiCloudStepFunctionCleanupByCron(Construct):
    def __init__(
        self,
        scope: Construct,
        construct_id: str,
        table: dynamodb.Table,
        aws_nuke_queue: sqs.Queue,
        azure_nuke_function: _lambda.Function,
        enviroment,
        **kwargs,
    ) -> None:
        super().__init__(scope, construct_id, **kwargs)

        query_expired_sandboxes_parallel = sfn.Parallel(self, "query_expired_sandboxes_parallel")

        """
        AWS Accounts
        """

        aws_dynamodb_query_expired_sandboxes_task = dynamodb_query_expired_sandboxes(
            self, f"{AWS_PREFIX} dynamodb_query_expired_sandboxes", table=table, cloud="aws"
        )
        query_expired_sandboxes_parallel.branch(aws_dynamodb_query_expired_sandboxes_task)

        map_aws = sfn.Map(self, f"{AWS_PREFIX} map", items_path="$.output.Items", result_path=sfn.JsonPath.DISCARD)

        aws_remove_sso_with_account_id = AWSRemoveWithID(
            self,
            f"{AWS_PREFIX} RemoveSSOWithAccountID",
            table=table,
            aws_nuke_queue=aws_nuke_queue,
            enviroment=enviroment,
        )

        map_aws.iterator(aws_remove_sso_with_account_id)
        aws_dynamodb_query_expired_sandboxes_task.next(map_aws)

        """
        AZURE Accounts
        """
        AZURE_PREFIX = "azure"
        azure_dynamodb_query_expired_sandboxes_task = dynamodb_query_expired_sandboxes(
            self, f"{AZURE_PREFIX} dynamodb_query_expired_sandboxes", table=table, cloud="azure"
        )
        query_expired_sandboxes_parallel.branch(azure_dynamodb_query_expired_sandboxes_task)

        map_azure = sfn.Map(self, f"{AZURE_PREFIX} map", items_path="$.output.Items", result_path=sfn.JsonPath.DISCARD)

        azure_remove_sso_with_account_id = AzureRemoveWithID(
            self, f"{AZURE_PREFIX} RemoveSSOWithAccountID", table=table, function=azure_nuke_function
        )

        map_azure.iterator(azure_remove_sso_with_account_id)

        azure_dynamodb_query_expired_sandboxes_task.next(map_azure)

        """
        Chain StepFunction
        """
        chain = sfn.Chain.start(query_expired_sandboxes_parallel)

        sfn_ = StepFunctionTemplate(self, "ProvisonSandbox", chain=chain).sfn

        dynamo_query = iam.PolicyStatement(
            actions=[
                "dynamodb:Query",
            ],
            resources=[table.table_arn, f"{table.table_arn}/*"],
        )
        sfn_.add_to_role_policy(dynamo_query)

        if enviroment == Enviroments.prod:
            sso_policy = iam.PolicyStatement(
                actions=["sso:ListAccountAssignments", "sso:DeleteAccountAssignment"],
                # TODO: add resources here
                resources=["*"],
            )
            sfn_.add_to_role_policy(sso_policy)
        self.step_function = sfn_


class MultiCloudStepFunctionCleanupByEvent(Construct):
    def __init__(
        self,
        scope: Construct,
        construct_id: str,
        table: dynamodb.Table,
        aws_nuke_queue: sqs.Queue,
        azure_nuke_function: _lambda.Function,
        enviroment,
        **kwargs,
    ) -> None:
        super().__init__(scope, construct_id, **kwargs)

        aws_remove = AWSRemoveWithID(
            self, f"{AWS_PREFIX} AWSRemoveWithID", table=table, aws_nuke_queue=aws_nuke_queue, enviroment=enviroment
        )

        azure_remove = AzureRemoveWithID(
            self, f"{AZURE_PREFIX} AzureRemoveWithID", table=table, function=azure_nuke_function
        )

        get_item = tasks.DynamoGetItem(
            self,
            "get item",
            table=table,
            key={
                "assigned_to": tasks.DynamoAttributeValue.from_string(sfn.JsonPath.string_at("$.detail.user")),
                "id": tasks.DynamoAttributeValue.from_string(sfn.JsonPath.string_at("$.detail.id")),
            },
            output_path="$.Item",
        )

        choice = sfn.Choice(self, "if AWS or Azure Event")

        choice.when(sfn.Condition.string_equals("$.cloud.S", "aws"), aws_remove)
        choice.when(sfn.Condition.string_equals("$.cloud.S", "azure"), azure_remove)

        choice.otherwise(sfn.Fail(self, "fail-state"))

        chain = sfn.Chain.start(get_item).next(choice)

        sfn_ = StepFunctionTemplate(self, "ProvisonSandbox", chain=chain).sfn

        if enviroment == Enviroments.prod:
            sso_policy = iam.PolicyStatement(
                actions=["sso:ListAccountAssignments", "sso:DeleteAccountAssignment"],
                # TODO: add resources here
                resources=["*"],
            )
            sfn_.add_to_role_policy(sso_policy)

        self.step_function = sfn_