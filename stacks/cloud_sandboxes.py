from typing import List
import os
from constructs import Construct
from aws_cdk import (
    Duration,
    Stack,
    aws_iam as iam,
    aws_sqs as sqs,
    aws_lambda as lambda_,
    aws_lambda_event_sources as lambda_event_sources,
    aws_lambda_python_alpha as lambda_python,
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

from stacks.sso_handler_cross_role import SSOHandlerCrossRole


class NukeLambdaWithQueue(Construct):
    def __init__(
        self,
        scope: Construct,
        id: str,
        asset: str,
        role: SSOHandlerCrossRole,
        **kwargs,
    ) -> None:
        super().__init__(scope, id, **kwargs)

        duration = Duration.minutes(5)

        nuke_queue = sqs.Queue(self, "nuke_queue", visibility_timeout=duration)
        nuke_lambda = lambda_python.PythonFunction(
            self,
            "nuke_lambda",
            architecture=lambda_.Architecture.ARM_64,
            runtime=lambda_.Runtime.PYTHON_3_9,
            entry=asset,
            index="handler.py",
            handler="handler",
            environment={
                "JUMP_ROLE_ARN": role.arn,
                "JUMP_ROLE_NAME": role.role_name,
                "LOGLEVEL": "INFO",
                "DRY_RUN": "True",
            },
            timeout=duration,
            memory_size=128,
            tracing=lambda_.Tracing.ACTIVE,
            insights_version=lambda_.LambdaInsightsVersion.from_insight_version_arn(
                "arn:aws:lambda:eu-central-1:580247275435:layer:LambdaInsightsExtension-Arm64:2"
            ),
        )
        nuke_lambda.add_event_source(lambda_event_sources.SqsEventSource(nuke_queue))

        policy = iam.PolicyStatement(
            actions=["sts:AssumeRole"],
            # TODO: add resources here dynamic
            resources=["arn:aws:iam::172920935848:role/nukeJumpRole-ssosandboxroleC6FC6E16-191Q1CC7HKKGR"],
        )
        nuke_lambda.add_to_role_policy(policy)

        self._lambda = nuke_lambda
        self.queue = nuke_queue


class PythonLambda(Construct):
    def __init__(
        self,
        scope: Construct,
        id: str,
        asset: str,
        **kwargs,
    ) -> None:
        super().__init__(
            scope,
            id,
            **kwargs,
        )
        duration = Duration.minutes(5)
        _lambda = lambda_python.PythonFunction(
            self,
            "lambda",
            architecture=lambda_.Architecture.ARM_64,
            runtime=lambda_.Runtime.PYTHON_3_9,
            entry=asset,
            index="handler.py",
            handler="handler",
            environment={
                "LOGLEVEL": "INFO",
                # "GITLAB_AZURE_PIPELINE_WEBHOOK": os.getenv("GITLAB_AZURE_PIPELINE_WEBHOOK", "NA"),
            },
            timeout=duration,
            memory_size=128,
            tracing=lambda_.Tracing.ACTIVE,
            insights_version=lambda_.LambdaInsightsVersion.from_insight_version_arn(
                "arn:aws:lambda:eu-central-1:580247275435:layer:LambdaInsightsExtension-Arm64:2"
            ),
        )
        self._lambda = _lambda


class CloudSandboxes(Stack):
    def __init__(
        self,
        scope: Construct,
        id: str,
        role: SSOHandlerCrossRole,
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
        # Lambda AZURE Add and Remove Api-Proxy
        # # # # # # # # # # # #
        azure_api_proxy = PythonLambda(self, "azure_api_proxy", "lambda/azure_api_proxy")

        # # # # # # # # # # # #
        # Cleanup Queue with Lambda: AWS
        # # # # # # # # # # # #
        aws_nuke_lambda = NukeLambdaWithQueue(self, "aws nukeLambda", asset="lambda/cleanup_aws", role=role)

        # # # # # # # # # # # #
        # Add by Event: AWS
        # # # # # # # # # # # #
        add_aws_step_function = AwsStepFunctionAdd(
            self,
            "AwsAddByEvent",
            table=multi_cloud_table.table,
            enviroment=enviroment,
        )

        eventHub.addRuleWithStepFunctionTarget(
            id="AwsAddByEventRule",
            sfnStepFunction=add_aws_step_function.step_function,
            detail={"user": [{"exists": True}], "action": ["add"], "cloud": ["aws"]},
        )

        # # # # # # # # # # # #
        # Add by Event: AZURE
        # # # # # # # # # # # #
        azure_aws_step_function = AzureStepFunctionAdd(
            self,
            "AzureAddByEvent",
            table=multi_cloud_table.table,
            function=azure_api_proxy._lambda,
            enviroment=enviroment,
        )

        eventHub.addRuleWithStepFunctionTarget(
            id="AzureAddByEventRule",
            sfnStepFunction=azure_aws_step_function.step_function,
            detail={"user": [{"exists": True}], "action": ["add"], "cloud": ["azure"]},
        )

        # # # # # # # # # # # #
        # Remove by Event: AWS / AZURE
        # # # # # # # # # # # #
        cleanup_event = MultiCloudStepFunctionCleanupByEvent(
            self,
            "MultiCloudCleanupByEvent",
            table=multi_cloud_table.table,
            aws_nuke_queue=aws_nuke_lambda.queue,
            azure_nuke_function=azure_api_proxy._lambda,
            enviroment=enviroment,
        )

        eventHub.addRuleWithStepFunctionTarget(
            id="MultiCloudCleanupByEventRule",
            sfnStepFunction=cleanup_event.step_function,
            detail={"user": [{"exists": True}], "action": ["remove"]},
        )

        # # # # # # # # # # # #
        # Cronjob to-Cleanup Sandboxes: AWS / AZURE
        # # # # # # # # # # # #
        cleanup_cron = MultiCloudStepFunctionCleanupByCron(
            self,
            "MultiCloudCleanupByCron",
            table=multi_cloud_table.table,
            aws_nuke_queue=aws_nuke_lambda.queue,
            azure_nuke_function=azure_api_proxy._lambda,
            enviroment=enviroment,
        )
        EventHubCron(self, "MultiCloudCleanupByCronRule", sfnStepFunction=cleanup_cron.step_function)

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