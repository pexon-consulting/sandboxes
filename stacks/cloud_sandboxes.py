from typing import List
from constructs import Construct
from aws_cdk import (
    Duration,
    Stack,
    aws_iam as iam,
    aws_sqs as sqs,
    aws_lambda as lambda_,
    aws_lambda_event_sources as lambda_event_sources,
)
from event_bus.infrastructure import EventHub, EventHubCron

from variables import sandboxes, root_account, region, Enviroments

from database.infrastructure import AWSTable
from hosting.infrastructure import AWSSandBoxHosting
from api.infrastructure import GraphQLEndpoint
from step_function.infrastructure import (
    AwsStepFunctionAdd,
    AzureStepFunctionAdd,
    MultiCloudStepFunctionCleanupByCron,
    MultiCloudStepFunctionCleanupByEvent,
)

from sso_handler.infrastructure import SSOHandler, SandboxGarbageCollector

from stacks.nuke_handler_cross_role import NukeHandlerCrossRole
from stacks.sso_handler_cross_role import SSOHandlerCrossRole


class NukeLambdaWithQueue(Construct):
    def __init__(
        self,
        scope: Construct,
        id: str,
        asset: str,
        **kwargs,
    ) -> None:
        super().__init__(scope, id, **kwargs)

        duration = Duration.minutes(5)

        nuke_queue = sqs.Queue(self, "nuke_queue", visibility_timeout=duration)
        nuke_lambda = lambda_.Function(
            self,
            "nuke_lambda",
            architecture=lambda_.Architecture.ARM_64,
            runtime=lambda_.Runtime.PYTHON_3_9,
            code=lambda_.Code.from_asset(asset),
            handler="handler.handler",
            environment={},
            timeout=duration,
            memory_size=128,
        )
        nuke_lambda.add_event_source(lambda_event_sources.SqsEventSource(nuke_queue))

        self._lambda = nuke_lambda
        self.queue = nuke_queue


class CloudSandboxes(Stack):
    def __init__(
        self,
        scope: Construct,
        id: str,
        # nuke_roles: List[NukeHandlerCrossRole],
        enviroment: Enviroments,
        **kwargs,
    ) -> None:
        super().__init__(scope, id, **kwargs)

        """
        Database
        """
        multi_cloud_table = AWSTable(self, "MultiCloudTable")

        """
        event-System
        """
        eventHub = EventHub(self, "EventBus")

        # # # # # # # # # # # #
        # Cleanup Queue with Lambda: AWS / AZURE
        # # # # # # # # # # # #
        aws_nuke_lambda = NukeLambdaWithQueue(self, "aws nukeLambda", asset="lambda/cleanup_aws")
        azure_nuke_lambda = NukeLambdaWithQueue(self, "azure nukeLambda", asset="lambda/cleanup_azure")

        # # # # # # # # # # # #
        # Add by Event: AWS
        # # # # # # # # # # # #
        add_aws_step_function = AwsStepFunctionAdd(
            self,
            "SsoStepFunction",
            table=multi_cloud_table.table,
            enviroment=enviroment,
        )

        eventHub.addRuleWithStepFunctionTarget(
            id="AddAWSSandbox",
            sfnStepFunction=add_aws_step_function.step_function,
            detail={"user": [{"exists": True}], "action": ["add"], "cloud": ["aws"]},
        )

        # # # # # # # # # # # #
        # Add by Event: AZURE
        # # # # # # # # # # # #
        azure_aws_step_function = AzureStepFunctionAdd(
            self,
            "AzureStepFunction",
            table=multi_cloud_table.table,
            enviroment=enviroment,
        )

        eventHub.addRuleWithStepFunctionTarget(
            id="AddAzureSandbox",
            sfnStepFunction=azure_aws_step_function.step_function,
            detail={"user": [{"exists": True}], "action": ["add"], "cloud": ["azure"]},
        )

        # # # # # # # # # # # #
        # Remove by Event: AWS / AZURE
        # # # # # # # # # # # #
        cleanup_event = MultiCloudStepFunctionCleanupByEvent(
            self,
            "RemoveSsoStepFunction",
            table=multi_cloud_table.table,
            aws_nuke_queue=aws_nuke_lambda.queue,
            azure_nuke_queue=azure_nuke_lambda.queue,
            enviroment=enviroment,
        )

        eventHub.addRuleWithStepFunctionTarget(
            id="RemoveAWSSandbox",
            sfnStepFunction=cleanup_event.step_function,
            detail={"user": [{"exists": True}], "action": ["remove"]},
        )

        # # # # # # # # # # # #
        # Cronjob to-Cleanup Sandboxes: AWS / AZURE
        # # # # # # # # # # # #
        cleanup_cron = MultiCloudStepFunctionCleanupByCron(
            self,
            "SsoStepFunctionCronCleanUp",
            table=multi_cloud_table.table,
            aws_nuke_queue=aws_nuke_lambda.queue,
            azure_nuke_queue=azure_nuke_lambda.queue,
            enviroment=enviroment,
        )
        EventHubCron(self, "EventHubCron", sfnStepFunction=cleanup_cron.step_function)

        """
        Graphql-Endpoint Backend
        """
        lambda_go_graphql = GraphQLEndpoint(
            self,
            "GraphQLEndpoint",
            multi_cloud_table=multi_cloud_table.table,
            eventHub=eventHub,
            enviroment=enviroment,
        )
        eventHub.bus.grant_all_put_events(lambda_go_graphql.func)

        self.functions = [lambda_go_graphql.func]
