from distutils import util
import logging
import os
import subprocess

logger = logging.getLogger()


def cloud_nuke(credentials) -> bool:
    lambda_key = os.getenv("AWS_ACCESS_KEY_ID")

    env = {
        "AWS_ACCESS_KEY_ID": credentials["AccessKeyId"],
        "AWS_SECRET_ACCESS_KEY": credentials["SecretAccessKey"],
        "AWS_SESSION_TOKEN": credentials["SessionToken"],
    }

    try:
        dry_run = util.strtobool(os.getenv("DRY_RUN", "True").lower())
    except Exception as e:
        logger.error(f'Fail to parse env-var "DRY_RUN" not a bool value')
        raise

    command = "--dry-run" if dry_run else "--force"
    cloud_nuke_tool = os.getenv("cloud_nuke", "cloud-nuke")
    LOG_LEVEL = os.getenv("LOG_LEVEL", "info")

    try:
        if lambda_key != env["AWS_ACCESS_KEY_ID"]:
            logger.debug("run")
            subprocess.check_call(
                [
                    f"./{cloud_nuke_tool} aws {command} --log-level {LOG_LEVEL} --region eu-central-1 --config ./cloud_nuke_config.yaml"
                ],
                shell=True,
                env=env,
            )
        return True

    except Exception as e:
        logger.error(f"Fail to call cloud-nuke: {e}")
        raise
