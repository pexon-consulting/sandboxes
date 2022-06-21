import builtins
import typing
from aws_cdk import Duration, aws_events as events, aws_events_targets as events_targets, aws_stepfunctions as sfn


from constructs import Construct

SOURCE = "de.pexon.sso.app"


class EventHub(Construct):
    def __init__(self, scope: Construct, construct_id: str, **kwargs) -> None:
        super().__init__(scope, construct_id, **kwargs)

        bus = events.EventBus(self, "bus")

        # events.Rule(
        #     self,
        #     "event_action_add",
        #     event_bus=bus,
        #     targets=[events_targets.SfnStateMachine(sfnStepFunction)],
        #     event_pattern={
        #         "source": [SOURCE],
        #         "detail": {"user": [{"exists": True}], "action": ["add"]},
        #     },
        # )

        # events.Rule(
        #     self,
        #     "azure_remove",
        #     event_bus=bus,
        #     targets=[events_targets.SfnStateMachine(sfnStepFunction)],
        #     event_pattern={
        #         "source": [SOURCE],
        #         "detail": {"pipelineID": [{"exists": True}], "action": ["remove"], "cloud": ["azure"]},
        #     },
        # )

        self.SOURCE = SOURCE
        self.bus = bus

    def addRuleWithStepFunctionTarget(
        self, id: str, sfnStepFunction: sfn.StateMachine, detail: typing.Mapping[builtins.str, typing.Any]
    ):
        events.Rule(
            self,
            f"event_rule_{id}",
            event_bus=self.bus,
            targets=[events_targets.SfnStateMachine(sfnStepFunction)],
            event_pattern={
                "source": [self.SOURCE],
                "detail": detail,
            },
        )


class EventHubCron(Construct):
    def __init__(self, scope: Construct, construct_id: str, sfnStepFunction: sfn.StateMachine, **kwargs) -> None:
        super().__init__(scope, construct_id, **kwargs)
        events.Rule(
            self,
            "ScheduleRule",
            schedule=events.Schedule.rate(Duration.days(1)),
            targets=[events_targets.SfnStateMachine(sfnStepFunction)],
        )
