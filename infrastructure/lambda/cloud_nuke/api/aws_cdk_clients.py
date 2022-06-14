import logging
import boto3

sts_client = boto3.client("sts")


def assume_role(RoleArn: str):
    assume_role = sts_client.assume_role(
        RoleArn=RoleArn,
        # Todo make this random
        RoleSessionName="cloud-nuke-lambda",
    )
    credentials = assume_role["Credentials"]
    return credentials


def s3_api(RoleArn: str):

    credentials = assume_role(RoleArn)
    session = boto3.Session(
        aws_access_key_id=credentials["AccessKeyId"],
        aws_secret_access_key=credentials["SecretAccessKey"],
        aws_session_token=credentials["SessionToken"],
    )
    return session.client("s3"), credentials
