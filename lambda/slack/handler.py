#!/usr/bin/python3.6
import urllib3
import json
import os
import logging
from aws_xray_sdk.core import xray_recorder
from aws_xray_sdk.core import patch_all

import requests

patch_all(double_patch=True)

http = urllib3.PoolManager()
slack_webhook = os.getenv("SLACK_WEBHOOK")
url = slack_webhook

LOGLEVEL = os.environ.get("LOGLEVEL", "INFO").upper()
logger = logging.getLogger()
logger.setLevel(level=LOGLEVEL)


def handler(event, context):
    print(event)
    try:
        msg_raw = event["Records"][0]["Sns"]["Message"]
        msg = json.loads(msg_raw)
        alarm_name = msg["AlarmName"]
        alarm_description = msg["AlarmDescription"]
        state_change_time = msg["StateChangeTime"]

    except:
        payload = {
            "blocks": [
                {
                    "type": "header",
                    "text": {"type": "plain_text", "text": "🚨🚨🚨 SLACK_WEBHOOK_ERROR 🚨🚨🚨", "emoji": True},
                },
            ]
        }
        requests.post(url, json=payload)
        return

    payload = {
        "blocks": [
            {"type": "header", "text": {"type": "plain_text", "text": "🚨 error occurred ⚠️", "emoji": True}},
            {
                "type": "section",
                "text": {
                    "type": "mrkdwn",
                    "text": "*Type:*\nALARM",
                },
            },
            {
                "type": "section",
                "text": {
                    "type": "mrkdwn",
                    "text": f"*AlarmName:*\n{alarm_name}",
                },
            },
            {
                "type": "section",
                "text": {
                    "type": "mrkdwn",
                    "text": f"*AlarmDescription:*\n{alarm_description}",
                },
            },
            {
                "type": "section",
                "text": {
                    "type": "mrkdwn",
                    "text": f"*StateChangeTime:*\n{state_change_time}",
                },
            },
        ]
    }

    requests.post(url, json=payload)
    return