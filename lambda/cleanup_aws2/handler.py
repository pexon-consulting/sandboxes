from json import JSONDecoder
import json
import logging
import os
import boto3

from aws_xray_sdk.core import xray_recorder
from aws_xray_sdk.core import patch_all
from aws_xray_lambda_segment_shim import get_sqs_triggered_recorder


from api.aws_cdk_clients import assume_role, client_from_credentials

from utils.cloud_nuke import cloud_nuke

LOGLEVEL = os.environ.get("LOGLEVEL", "INFO").upper()
logger = logging.getLogger()
logger.setLevel(level=LOGLEVEL)

patch_all()


def handler(event=None, context=None):
    if event["Records"] and len(event["Records"]) > 0:
        for i, record in enumerate(event["Records"]):
            recorder = get_sqs_triggered_recorder(
                record=record,
                lambda_request_id=context.aws_request_id,
                lambda_arn=context.invoked_function_arn,
            )
            with recorder.in_segment():
                with recorder.in_subsegment(f"SQS Record {i}") as subsegment:
                    print(
                        "I'm triggered by an SQS Record and using trace id ",
                        subsegment.trace_id,
                    )
    logger.info(event)
    JUMP_ROLE_NAME = os.getenv("JUMP_ROLE_NAME")

    if event["Records"] and len(event["Records"]) > 0:
        for record in event["Records"]:
            body_raw = record["body"]
            body = json.loads(body_raw)
            sandbox_account_id = body["account_id"]
            try:
                logger.info("try to assume **nukeJumpRole** role")
                credentials_nuke_jump_role = assume_role(
                    # Replace role name and make it dynamic
                    "arn:aws:iam::172920935848:role/nukeJumpRole-ssosandboxroleC6FC6E16-191Q1CC7HKKGR",
                    boto3.client("sts"),
                )
                logger.info("successful assumed **nukeJumpRole** role")
            except Exception as e:
                logger.exception("fail to assume **nukeJumpRole** role")
                logger.exception(e)
                return {"statusCode": 400}

            try:
                logger.info(f"try to assume **OrganizationAccountAccessRole** role from account {sandbox_account_id}")
                credentials_sandbox = assume_role(
                    f"arn:aws:iam::{sandbox_account_id}:role/OrganizationAccountAccessRole",
                    client_from_credentials("sts", credentials_nuke_jump_role),
                )
                logger.info(
                    f"successful assumed **OrganizationAccountAccessRole** role  from account {sandbox_account_id}"
                )
            except Exception as e:
                logger.exception(
                    f"fail to assume **OrganizationAccountAccessRole** role  from account {sandbox_account_id}"
                )
                logger.exception(e)
                return {"statusCode": 400}

            try:
                cloud_nuke(credentials_sandbox)
            except Exception as e:
                logger.exception(e)
                logger.exception(f"fail to nuke {sandbox_account_id}")
                return {"statusCode": 400}

    return {"statusCode": 200}
