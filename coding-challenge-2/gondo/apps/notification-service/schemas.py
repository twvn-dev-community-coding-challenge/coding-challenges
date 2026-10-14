"""HTTP request bodies (transport layer)."""

from __future__ import annotations

from pydantic import BaseModel, Field


class SmsChannelPayload(BaseModel):
    country_code: str
    phone_number: str


class CreateNotificationRequest(BaseModel):
    message_id: str
    channel_type: str = "SMS"
    recipient: str
    content: str
    channel_payload: SmsChannelPayload
    issue_server_otp: bool = Field(
        default=False,
        description="If true, call otp-service to generate a code (hashed at rest with TTL); "
        "substitutes {{OTP}} in content.",
    )


class DispatchRequest(BaseModel):
    as_of: str = Field(..., description="ISO 8601 datetime string")


class RetryRequest(BaseModel):
    as_of: str | None = None


class ProviderCallbackRequest(BaseModel):
    message_id: str
    provider: str
    new_state: str
    actual_cost: float | None = Field(
        default=None,
        description=(
            "Billed amount for charging-service RecordActualCost. "
            "Send-success: optional (falls back to estimated_cost). "
            "Send-failed: omit to skip RecordActualCost."
        ),
    )
