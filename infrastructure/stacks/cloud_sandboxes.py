from typing import List
from constructs import Construct
from aws_cdk import Stack, aws_iam as iam, aws_sqs as sqs
from event_bus.infrastructure import EventHub, EventHubCron

from variables import sandboxes, root_account, region, Enviroments

from database.infrastructure import AWSTable
from hosting.infrastructure import AWSSandBoxHosting
from api.infrastructure import GraphQLEndpoint
from sso_step_function.infrastructure import SsoStepFunctionAdd, SsoStepFunctionCronCleanUp

from sso_handler.infrastructure import SSOHandler, SandboxGarbageCollector

from stacks.nuke_handler_cross_role import NukeHandlerCrossRole
from stacks.sso_handler_cross_role import SSOHandlerCrossRole


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
        aws_nuke_queue = sqs.Queue(self, "aws_nuke_queue")
        azure_nuke_queue = sqs.Queue(self, "azure_nuke_queue")

        sso_step_function = SsoStepFunctionAdd(
            self,
            "SsoStepFunction",
            table=multi_cloud_table.table,
            enviroment=enviroment,
        )
        eventHub = EventHub(self, "EventBus", sfnStepFunction=sso_step_function.step_function)

        sso_step_cron = SsoStepFunctionCronCleanUp(
            self,
            "SsoStepFunctionCronCleanUp",
            table=multi_cloud_table.table,
            aws_nuke_queue=aws_nuke_queue,
            enviroment=enviroment,
        )

        EventHubCron(self, "EventHubCron", sfnStepFunction=sso_step_cron.step_function)

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
        AWSSandBoxHosting(
            self,
            "Hosting",
            ssm_sandbox_domain_uri=lambda_go_graphql.ssm_sandbox_domain_uri,
            enviroment=enviroment,
        )

        self.functions = [lambda_go_graphql.func]
