from aws_cdk import Duration, aws_events as events, aws_events_targets as events_targets, aws_stepfunctions as sfn


from constructs import Construct


class EventHub(Construct):
    def __init__(self, scope: Construct, construct_id: str, sfnStepFunction: sfn.StateMachine, **kwargs) -> None:
        super().__init__(scope, construct_id, **kwargs)
        source = "de.pexon.sso.app"
        bus = events.EventBus(self, "bus")

        events.Rule(
            self,
            "event_action_add",
            event_bus=bus,
            targets=[events_targets.SfnStateMachine(sfnStepFunction)],
            event_pattern={
                "source": [source],
                "detail": {"user": [{"exists": True}], "action": ["add"]},
            },
        )

        self.source = source
        self.bus = bus


class EventHubCron(Construct):
    def __init__(self, scope: Construct, construct_id: str, sfnStepFunction: sfn.StateMachine, **kwargs) -> None:
        super().__init__(scope, construct_id, **kwargs)
        events.Rule(
            self,
            "ScheduleRule",
            schedule=events.Schedule.rate(Duration.days(1)),
            targets=[events_targets.SfnStateMachine(sfnStepFunction)],
        )
