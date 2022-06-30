#!/usr/bin/python3.6
# {'id': '3812b9e5-f5a7-4f37-b15c-293a57344846', 'user': 'azure.test@pexon-consulting.de', 'action': 'add', 'assigned_until': '2022-07-02T00:00:00Z', 'assigned_since': '2022-06-30T13:55:16Z', 'cloud': 'aws'}
import os
import requests
from requests.structures import CaseInsensitiveDict


def handler(event, context):
    if event["action"] == "add":
        user: str = event["user"]

        state_name = user.replace(".", "-").split("@")[0]
        sandbox_name = f"rg-bootcamp-{state_name}"

        url = os.getenv("GITLAB_AZURE_PIPELINE_WEBHOOK", "https://httpbin.org/post")

        headers = CaseInsensitiveDict()
        headers["Content-Type"] = "application/x-www-form-urlencoded"

        data = f"rg_name={sandbox_name}&trainee_email={user}&created_by={user}&variables[TF_STATE_NAME]={state_name}"
        resp = requests.post(url, headers=headers, data=data)

        print(resp.json())
        # TODO return here a map with ids to dump into dynamoDB
        return {"pipeline-id": "123123123", "azure-magic": "12312313"}

    if event["action"] == "remove":
        # dome some remove action here
        pass
    return {"statusCode": 400}
