import logging
from api.aws_cdk_clients import s3_api
from utils.cloud_nuke import cloud_nuke
from utils.clear_bucket_policy import clear_bucket_policy

logging.basicConfig(level=logging.ERROR)


def handler(event=None, context=None):

    role = f"arn:aws:iam::815837829183:role/prod-sandbox-4dboxrolesandbox415420752fec06db2dea9"

    s3_client, credentials_account_to_nuke = s3_api(role)

    clear_bucket_policy(s3_client)

    cloud_nuke(credentials_account_to_nuke)

    return {"statusCode": 200}
