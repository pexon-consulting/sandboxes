import json
import logging
import boto3
from sso_assignment import create_sso_assigment, remove_sso_assigment_and_nuke

logging.basicConfig(level=logging.ERROR)
# create
# delete
x = {
    "Records": [
        {
            "messageId": "0ae74e6c-285f-4c15-9ca8-4e9991a97773",
            "receiptHandle": "AQEBvYwkrukiGPwxQipDdji5P1leqxkX4W0+pxTC9CY8zJIovTJatcRPH1t7+C7iigyjnvKOVEBLl4TSSFNZNXadYxMWV4bclDQQPdHUMiQF55BfL50z286B6LEWoAqOpJwahn073+DKG/wtw2/1R046XVgP2Z1yKJKpezc+x94cmUXUiCcHpr0XZShGtLNnDPJPzu0oF5whYpXDPPflGFXesJsxgJf5CKWIQaBgG1qUV5tRGAshrtix5bd3tF1EGDmzHlM7B7QuxYLkXKqW2gRbFhd3Q48KB8B0D/EZIfP5EcksPFM2ZOtb73n6FkczIXTnLwhfP2x5J/XPbZcTgSNOeyWoTAxVymiPjVbiXWsXdYRPK8DHISnGG3xzigjb5jLvKPKuH4DwMLEtDTw00PYLD+V4b95w+vWRUwedbyD9GocqK/z6N66vDJi6vAR4I6tw",
            "body": '{\n"event":"delete"\n,\n"assigned_to":"maximilian.haensel@pexon-consulting.de"\n,\n"account_id":"815837829183"\n,\n"sandbox":"test"\n}',
            "attributes": {
                "ApproximateReceiveCount": "1",
                "SentTimestamp": "1649966784289",
                "SenderId": "AROAQ5UUE7XW3H3CU4TWZ:maximilian.haensel@pexon-consulting.de",
                "ApproximateFirstReceiveTimestamp": "1649966784290",
            },
            "messageAttributes": {},
            "md5OfBody": "2a06af0a7930b36836d1f977e3f79460",
            "eventSource": "aws:sqs",
            "eventSourceARN": "arn:aws:sqs:eu-central-1:063661473261:prod-Sandbox-SSOHandlerQueue99D7B3B6-cOq9xwfV4nkD",
            "awsRegion": "eu-central-1",
        }
    ]
}


def event_processor(body):

    event = body["event"]
    assigned_to = body["assigned_to"]
    account_id = body["account_id"]
    sandbox = body["sandbox"]

    if event == "create":
        create_sso_assigment(assigned_to, account_id)
    if event == "delete":
        remove_sso_assigment_and_nuke(assigned_to, account_id)


def handler(event=None, context=None):
    for record in x["Records"]:
        body = json.loads(record["body"])
        event_processor(body)

    return {"statusCode": 200}
