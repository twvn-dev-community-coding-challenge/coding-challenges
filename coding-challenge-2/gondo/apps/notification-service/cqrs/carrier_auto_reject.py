"""Simulated MNO rejection for MSISDNs listed in ``mock_scenarios.json`` (``carrier_rejected``)."""

from __future__ import annotations

from cqrs.mock_scenario_loader import get_mock_outcome


def should_auto_carrier_reject(country_code: str | None, phone_number: str) -> bool:
    """True when the carrier layer simulates MNO rejection for this recipient."""
    return get_mock_outcome(phone_number, country_code) == "carrier_rejected"
