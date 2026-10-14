"""Dev-only optional mock: auto-complete Send-to-carrier → Send-success / Send-failed for test MSISDNs.

When ``MOCK_SMS_SUCCESS_PHONES`` is **set** to a non-empty comma-separated list, digits are
matched after normalization (legacy behavior). When the variable is **unset** or **empty**,
outcomes are read from ``mock_scenarios.json``.

For **VN**, national input such as ``0999999999`` (with ``country_code`` ``VN``) normalizes like
the membership UI: strip a leading ``0`` then prefix ``84``.

Normalization helpers live in ``cqrs.mock_scenario_loader`` (single source of truth).
"""

from __future__ import annotations

import logging
import os
from datetime import datetime, timezone

from cqrs.mock_scenario_loader import (
    canonical_mock_msisdn_digits,
    get_mock_actual_cost,
    get_mock_outcome,
    normalize_digits,
)

logger = logging.getLogger(__name__)

_ENV_KEY = "MOCK_SMS_SUCCESS_PHONES"


def _uses_env_success_allowlist() -> bool:
    """True when ``MOCK_SMS_SUCCESS_PHONES`` is present and non-empty."""
    if _ENV_KEY not in os.environ:
        return False
    return bool(os.environ.get(_ENV_KEY, "").strip())


def should_autocomplete_delivery_success(
    phone: str | None,
    country_code: str | None = None,
    attempt: int = 1,
) -> bool:
    """True when this recipient should mock-complete to Send-success (or retry success path)."""
    if _uses_env_success_allowlist():
        raw = os.environ[_ENV_KEY]
        allowed = {normalize_digits(x) for x in raw.split(",") if x.strip()}
        if not allowed:
            return False
        candidate = canonical_mock_msisdn_digits(country_code, phone)
        return bool(candidate) and candidate in allowed

    outcome = get_mock_outcome(phone, country_code)
    if outcome == "send_success":
        return True
    if outcome == "retry_then_success" and attempt > 1:
        return True
    return False


def _pipeline_reason_for_success_mock() -> str:
    return "MOCK_SMS_SUCCESS_PHONES" if _uses_env_success_allowlist() else "mock_scenarios.json"


async def apply_mock_send_success_if_eligible(message_id: str) -> None:
    """If notification is Send-to-carrier with a mock MSISDN, transition to Send-success."""
    from cqrs.charging_callbacks import record_actual_cost_after_callback
    from cqrs.transitions import record_transition
    from models import is_valid_transition
    from pipeline_runtime import append_pipeline_event
    from schemas import ProviderCallbackRequest
    from store import find_by_message_id, update_notification

    n = find_by_message_id(message_id)
    if n is None or n.state != "Send-to-carrier":
        return
    phone = n.channel_payload.get("phone_number", "")
    country_code = n.channel_payload.get("country_code", "")
    if not should_autocomplete_delivery_success(phone, country_code, attempt=n.attempt):
        return

    if _uses_env_success_allowlist():
        actual_cost: float | None = n.estimated_cost
    else:
        actual_cost = get_mock_actual_cost(phone, country_code)

    body = ProviderCallbackRequest(
        message_id=n.message_id,
        provider="mock-dev",
        new_state="Send-success",
        actual_cost=actual_cost,
    )
    if not is_valid_transition(n.state, body.new_state):
        return

    from_st = n.state
    n.state = "Send-success"
    n.updated_at = datetime.now(timezone.utc)
    update_notification(n)
    record_transition(
        n.notification_id,
        from_st,
        "Send-success",
        "system",
        "accepted",
        "mock_autocomplete_send_success",
    )
    append_pipeline_event(
        n.notification_id,
        "state.Send-success",
        {
            "from_state": from_st,
            "reason": _pipeline_reason_for_success_mock(),
            "phone_number": phone,
        },
    )
    logger.info(
        "mock_autocomplete_send_success",
        extra={
            "notification_id": n.notification_id,
            "message_id": n.message_id,
            "phone_number": phone,
        },
    )
    await record_actual_cost_after_callback(n, body)


def should_apply_mock_send_failed(phone: str | None, country_code: str | None) -> bool:
    """True when ``mock_scenarios.json`` maps this recipient to ``send_failed``."""
    if _uses_env_success_allowlist():
        return False
    return get_mock_outcome(phone, country_code) == "send_failed"


async def apply_mock_send_failed_if_eligible(message_id: str) -> None:
    """If still Send-to-carrier and scenario is ``send_failed``, transition to Send-failed."""
    from cqrs.charging_callbacks import record_actual_cost_after_callback
    from cqrs.transitions import record_transition
    from models import is_valid_transition
    from pipeline_runtime import append_pipeline_event
    from schemas import ProviderCallbackRequest
    from store import find_by_message_id, update_notification

    n = find_by_message_id(message_id)
    if n is None or n.state != "Send-to-carrier":
        return
    phone = n.channel_payload.get("phone_number", "")
    country_code = n.channel_payload.get("country_code", "")
    if not should_apply_mock_send_failed(phone, country_code):
        return

    body = ProviderCallbackRequest(
        message_id=n.message_id,
        provider="mock-dev",
        new_state="Send-failed",
        actual_cost=get_mock_actual_cost(phone, country_code),
    )
    if not is_valid_transition(n.state, body.new_state):
        return

    from_st = n.state
    n.state = "Send-failed"
    n.updated_at = datetime.now(timezone.utc)
    update_notification(n)
    record_transition(
        n.notification_id,
        from_st,
        "Send-failed",
        "system",
        "accepted",
        "mock_autocomplete_send_failed",
    )
    append_pipeline_event(
        n.notification_id,
        "state.Send-failed",
        {
            "from_state": from_st,
            "reason": "mock_scenarios.json",
            "phone_number": phone,
        },
    )
    logger.info(
        "mock_autocomplete_send_failed",
        extra={
            "notification_id": n.notification_id,
            "message_id": n.message_id,
            "phone_number": phone,
        },
    )
    await record_actual_cost_after_callback(n, body)
