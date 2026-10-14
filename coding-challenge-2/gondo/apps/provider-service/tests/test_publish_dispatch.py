"""publish_sms_dispatch_requested — optional CARRIER_HTTP_PROBE_URL override."""

from __future__ import annotations

from unittest.mock import AsyncMock, MagicMock, patch

import pytest

from cqrs.events import SmsDispatchRequested


@pytest.mark.asyncio
async def test_carrier_http_probe_url_overrides_yaml_endpoint(monkeypatch: pytest.MonkeyPatch) -> None:
    monkeypatch.setenv("CARRIER_HTTP_PROBE_URL", "http://wiremock:8080/carrier/sms")
    bus = MagicMock()
    bus.publish = AsyncMock()

    with (
        patch("cqrs.publish_dispatch.get_message_bus", return_value=bus),
        patch("cqrs.publish_dispatch.publish_json", new_callable=AsyncMock) as p_pub,
    ):
        from cqrs.publish_dispatch import publish_sms_dispatch_requested

        ok = await publish_sms_dispatch_requested(
            message_id="m1",
            country_code="VN",
            carrier="VIETTEL",
            selected_provider_id="prv_02",
            selected_provider_code="VONAGE",
        )

    assert ok is True
    p_pub.assert_awaited_once()
    _bus, _topic, event = p_pub.await_args[0]
    assert isinstance(event, SmsDispatchRequested)
    assert event.api_endpoint == "http://wiremock:8080/carrier/sms"
