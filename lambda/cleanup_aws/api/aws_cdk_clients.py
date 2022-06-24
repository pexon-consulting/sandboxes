import logging
import boto3


def client_from_credentials(client: str, credentials):
    return boto3.client(
        client,
        aws_access_key_id=credentials["AccessKeyId"],
        aws_secret_access_key=credentials["SecretAccessKey"],
        aws_session_token=credentials["SessionToken"],
    )


def assume_role(RoleArn: str, client):
    assume_role = client.assume_role(
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
