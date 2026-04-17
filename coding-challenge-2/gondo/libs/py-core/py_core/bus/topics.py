"""Logical topic names shared across services (see infra/message-bus/topics.yaml)."""

# Provider → Carrier: work item to send SMS via provider HTTP API
SMS_DISPATCH_REQUESTED = "sms.dispatch.requested"

# Carrier → Provider: carrier accepted the dispatch job (before outbound HTTP)
SMS_DISPATCH_RECEIVED = "sms.dispatch.received"

# Carrier → Provider (and others): delivery outcome (rate limits respected in carrier)
SMS_DISPATCH_OUTCOME = "sms.dispatch.outcome"


def topic_to_subject(topic: str, prefix: str = "gondo") -> str:
    """Map a logical topic to a NATS subject."""
    t = topic.strip().lstrip(".")
    return f"{prefix}.{t}"
