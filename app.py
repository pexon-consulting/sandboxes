#!/usr/bin/env python3

# For consistency with TypeScript code, `cdk` is the preferred import name for
# the CDK's core module.  The following line also imports it as `core` for use
# with examples from the CDK Developer's Guide, which are in the process of
# being updated to use `cdk`.  You may delete this import if you don't need it.
from typing import List
from aws_cdk import (
    App,
    Environment,
    Stage,
    Tags,
)

from stacks.cloud_sandboxes import CloudSandboxes
from stacks.sso_handler_cross_role import SSOHandlerCrossRole
from stacks.monitoring import Monitoring
from stacks.cicd_preperation import CICDPreperation

from variables import sandboxes, root_account, region, root_pexon, Enviroments

env_EU = Environment(account=root_account, region=region)

app = App()

Tags.of(app).add("needUntil", "2099-01-01T00:00:00.000Z")
Tags.of(app).add("creator", "maximilian.haensel@pexon-consulting.de")
Tags.of(app).add("app", "aws-sandbox-handler")


class SandboxStage(Stage):
    def __init__(self, scope, id: str, enviroment: Enviroments, role: SSOHandlerCrossRole, *, env=None, outdir=None):
        super().__init__(scope, id, env=env, outdir=outdir)

        Tags.of(self).add("stage", id)

        sandbox = CloudSandboxes(
            self,
            "Sandbox",
            role=role,
            enviroment=enviroment,
        )
        self.stack = sandbox


role = SSOHandlerCrossRole(
    app,
    "nukeJumpRole",
    env={
        "account": root_pexon,
        "region": region,
    },
)
cicd = CICDPreperation(app, "CICDPreperation", env=env_EU)

prod = SandboxStage(app, Enviroments.prod.value, env=env_EU, enviroment=Enviroments.prod, role=role)
test = SandboxStage(app, Enviroments.test.value, env=env_EU, enviroment=Enviroments.test, role=role)


app.synth()