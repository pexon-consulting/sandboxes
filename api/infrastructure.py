import os
from typing import List
from constructs import Construct
from aws_cdk import (
    Duration,
    aws_s3 as s3,
    aws_ssm as ssm,
    aws_lambda as lambda_,
    aws_apigateway as apigw,
    aws_dynamodb as dynamodb,
    aws_sqs as sqs,
    aws_lambda_go_alpha as go_lambda,
    aws_logs as log,
    # aws_apigatewayv2_authorizers_alpha as authorizers,
)
import jsii
from event_bus.infrastructure import EventHub

from variables import Enviroments


@jsii.implements(go_lambda.ICommandHooks)
class CommandHooks:
    def before_bundling(self, input_dir, output_dir):
        return ["go generate ./schema"]

    def after_bundling(self, input_dir, output_dir):
        return []


class GraphQLEndpoint(Construct):
    def __init__(
        self,
        scope: Construct,
        construct_id: str,
        multi_cloud_table: dynamodb.Table,
        eventHub: EventHub,
        enviroment: Enviroments,
    ) -> None:
        super().__init__(scope, construct_id)

        func = go_lambda.GoFunction(
            self,
            "lambda",
            entry=f"{os.getcwd()}/lambda/graph-ql-api",
            architecture=lambda_.Architecture.ARM_64,
            timeout=Duration.seconds(30),
            memory_size=128,
            environment={
                "DYNAMODB_TABLE": multi_cloud_table.table_name,
                "gitlab_azure_pipeline_webhook": os.getenv("GITLAB_AZURE_PIPELINE_WEBHOOK", "NA"),
                "event_source": eventHub.SOURCE,
                "event_bus_name": eventHub.bus.event_bus_name,
            },
            bundling=go_lambda.BundlingOptions(
                command_hooks=CommandHooks(),
            ),
            tracing=lambda_.Tracing.ACTIVE,
            insights_version=lambda_.LambdaInsightsVersion.from_insight_version_arn(
                "arn:aws:lambda:eu-central-1:580247275435:layer:LambdaInsightsExtension-Arm64:2"
            ),
            log_retention=log.RetentionDays.TWO_MONTHS,
        )

        multi_cloud_table.grant_read_write_data(func)

        api = apigw.LambdaRestApi(
            self,
            f"GraphqlApi-{enviroment.value}",
            handler=func,
            proxy=False,
            default_cors_preflight_options=apigw.CorsOptions(
                allow_origins=apigw.Cors.ALL_ORIGINS, allow_methods=apigw.Cors.ALL_METHODS
            ),
            deploy_options={
                "tracing_enabled": True,
                "stage_name": enviroment.value,
                "logging_level": apigw.MethodLoggingLevel.ERROR,
            },
        )

        ssm_sandbox_domain_uri = ssm.StringParameter(
            self,
            "SandboxDomainUri",
            description="Name of the API URI",
            parameter_name=f"/{enviroment.value}/sandbox/apiuri",
            string_value=api.url,
        )

        items = api.root.add_resource("graphql")
        items.add_method("POST")  # POST /sandbox

        self.func = func
        self.ssm_sandbox_domain_uri = ssm_sandbox_domain_uri
