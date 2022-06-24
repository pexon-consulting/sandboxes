from typing import List
from constructs import Construct
from aws_cdk import (
    Stack,
    PhysicalName,
    aws_iam as iam,
)

from variables import root_account


class SSOHandlerCrossRole(Stack):
    def __init__(self, scope: Construct, id: str, **kwargs) -> None:
        super().__init__(scope, id, **kwargs)
        role = iam.Role(
            self,
            f"sso-sandbox-role",
            role_name=PhysicalName.GENERATE_IF_NEEDED,
            assumed_by=iam.AccountPrincipal(root_account),
        )

        role.add_to_principal_policy(iam.PolicyStatement(actions=["*"], resources=["*"]))

        self.role_name = role.role_name
        self.id = id
        self.arn = role.role_arn
