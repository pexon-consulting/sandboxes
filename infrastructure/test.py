import json
from turtle import clear
import boto3
import uuid


# client = boto3.client("dynamodb")

# response = client.query(
#     TableName="test-Sandbox-MultiCloudTableMultiCloudSandboxF1DC88C9-1QO0PBD55LGGN",
#     IndexName="gsi_cloud",
#     Select="ALL_PROJECTED_ATTRIBUTES",
#     FilterExpression="aws.account_id = :account_id AND assigned_until > :assigned_until",
#     KeyConditionExpression="cloud = :cloud AND #order_state = :state",
#     ExpressionAttributeValues={
#         ":cloud": {"S": "aws"},
#         ":state": {"S": "accounted"},
#         ":account_id": {"S": "12345"},
#         ":assigned_until": {"S": "2022-06-01T00:00:00Z"},
#     },
#     ExpressionAttributeNames={"#order_state": "state"},
# )

# print(response["Items"])

session = boto3.Session(profile_name="sandbox")
client = session.client("events")

for _ in range(2):
    response = client.put_events(
        Entries=[
            {
                # "Time": datetime(2015, 1, 1),
                "Source": "de.pexon.sso.app",
                "Resources": [
                    "string",
                ],
                "DetailType": "string",
                "Detail": json.dumps(
                    {
                        "id": str(uuid.uuid4()),
                        "user": "maximilian.haensel@pexon-consulting.de",
                        "action": "add",
                        "assigned_until": "200",
                        "assigned_since": "300",
                    }
                ),
                "EventBusName": "testSandboxEventBusbus4E22DCFC",
                "TraceHeader": "string",
            },
        ],
    )
print(response)
