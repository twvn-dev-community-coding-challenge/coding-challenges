"""Dev-only optional mock: auto-complete Send-to-carrier → Send-success for test MSISDNs.

Set ``MOCK_SMS_SUCCESS_PHONES`` to a comma-separated list (digits compared after
normalization). Default when unset:
``+84999999999,+8490909090`` (canonical digits ``84999999999`` and ``8490909090``).

For **VN**, national input such as ``0999999999`` or ``090909090`` (with ``country_code``
``VN``) is normalized like the membership UI: strip a leading ``0`` then prefix ``84``,
so it matches the same digits as ``+84999999999`` / ``+8490909090``.

Set ``MOCK_SMS_SUCCESS_PHONES`` to an empty string to disable."""

from __future__ import annotations

import logging
import os
from datetime import datetime, timezone

logger = logging.getLogger(__name__)

_ENV_KEY = "MOCK_SMS_SUCCESS_PHONES"


def normalize_digits(phone: str | None) -> str:
    if not phone:
        return ""
    return "".join(c for c in phone.strip() if c.isdigit())


def canonical_mock_msisdn_digits(country_code: str | None, phone: str | None) -> str:
    """Single canonical digit string for mock comparison (aligns with UI ``buildSmsPhoneNumber`` for VN)."""
    d = normalize_digits(phone)
    if not d:
        return ""
    iso = (country_code or "").strip().upper()
    if iso == "VN":
        # National: leading 0 + subscriber digits → 84 + subscriber (matches membership UI)
        if d.startswith("0") and len(d) >= 2:
            return "84" + d[1:]
        # Nine subscriber digits without leading 0 (user omitted the national 0)
        if len(d) == 9 and not d.startswith("0"):
            return "84" + d
        # International form without +: 84 + national significant number
        if d.startswith("84") and len(d) >= 10:
            return d
    return d


def should_autocomplete_delivery_success(
    phone: str | None,
    country_code: str | None = None,
) -> bool:
    """True when canonical MSISDN digits match the configured mock allowlist."""
    if _ENV_KEY in os.environ:
        raw = os.environ.get(_ENV_KEY, "")
    else:
        raw = "+84999999999,+8490909090"
    if not raw.strip():
        return False
    allowed = {normalize_digits(x) for x in raw.split(",") if x.strip()}
    if not allowed:
        return False
    candidate = canonical_mock_msisdn_digits(country_code, phone)
    return bool(candidate) and candidate in allowed


async def apply_mock_send_success_if_eligible(message_id: str) -> None:
    """If notification is Send-to-carrier with a mock MSISDN, transition to Send-success."""
    from cqrs.charging_callbacks import record_actual_cost_after_callback
    from cqrs.transitions import record_transition
    from schemas import ProviderCallbackRequest
    from models import is_valid_transition
    from pipeline_runtime import append_pipeline_event
    from store import find_by_message_id, update_notification

    n = find_by_message_id(message_id)
    if n is None or n.state != "Send-to-carrier":
        return
    phone = n.channel_payload.get("phone_number", "")
    country_code = n.channel_payload.get("country_code", "")
    if not should_autocomplete_delivery_success(phone, country_code):
        return

    body = ProviderCallbackRequest(
        message_id=n.message_id,
        provider="mock-dev",
        new_state="Send-success",
        actual_cost=n.estimated_cost,
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
            "reason": "MOCK_SMS_SUCCESS_PHONES",
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
