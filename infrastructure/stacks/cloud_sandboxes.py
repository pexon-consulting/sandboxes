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
from sso_step_function.infrastructure import SsoStepFunctionAdd, SsoStepFunctionCronCleanUp, SsoStepFunctionRemove

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
        nuke_lambda.add_event_source(
            lambda_event_sources.SqsEventSource(nuke_queue))

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

        aws_nuke_lambda = NukeLambdaWithQueue(
            self, "aws nukeLambda", asset="lambda/cloud_nuke")
        azure_nuke_lambda = NukeLambdaWithQueue(
            self, "azure nukeLambda", asset="lambda/azure_nuke")

        sso_step_function = SsoStepFunctionAdd(
            self,
            "SsoStepFunction",
            table=multi_cloud_table.table,
            enviroment=enviroment,
        )

        eventHub.addRuleWithStepFunctionTarget(
            id="AddAWSSandbox",
            sfnStepFunction=sso_step_function.step_function,
            detail={"user": [{"exists": True}], "action": ["add"]},
        )

        azure_step_function = SsoStepFunctionAdd(
            self,
            "AzureStepFunction",
            table=multi_cloud_table.table,
            enviroment=enviroment,
        )

        eventHub.addRuleWithStepFunctionTarget(
            id="AddAzureSandbox",
            sfnStepFunction=azure_step_function.step_function,
            detail={"user": [{"exists": True}], "action": ["add"]},
        )

        step_function_remove = SsoStepFunctionRemove(
            self,
            "RemoveSsoStepFunction",
            table=multi_cloud_table.table,
            aws_nuke_queue=aws_nuke_lambda.queue,
            enviroment=enviroment,
        )

        eventHub.addRuleWithStepFunctionTarget(
            id="RemoveAWSSandbox",
            sfnStepFunction=step_function_remove.step_function,
            detail={"user": [{"exists": True}], "action": ["remove"]},
        )

        sso_step_cron = SsoStepFunctionCronCleanUp(
            self,
            "SsoStepFunctionCronCleanUp",
            table=multi_cloud_table.table,
            aws_nuke_queue=aws_nuke_lambda.queue,
            azure_nuke_queue=azure_nuke_lambda.queue,
            enviroment=enviroment,
        )
        EventHubCron(self, "EventHubCron",
                     sfnStepFunction=sso_step_cron.step_function)

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
        """
        create Web-App-Hosting 
        """
        # AWSSandBoxHosting(
        #     self,
        #     "Hosting",
        #     ssm_sandbox_domain_uri=lambda_go_graphql.ssm_sandbox_domain_uri,
        #     enviroment=enviroment,
        # )

        self.functions = [lambda_go_graphql.func]
