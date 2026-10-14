"""Load ``mock_scenarios.json`` and resolve SMS mock outcomes by normalized MSISDN."""

from __future__ import annotations

import json
import logging
from pathlib import Path
from typing import Any

logger = logging.getLogger(__name__)

_MOCK_SCENARIOS_PATH = Path(__file__).resolve().parent.parent / "mock_scenarios.json"


def normalize_digits(phone: str | None) -> str:
    """Strip a phone string to digits only."""
    if not phone:
        return ""
    return "".join(c for c in phone.strip() if c.isdigit())


def canonical_mock_msisdn_digits(country_code: str | None, phone: str | None) -> str:
    """Canonical digit string for mock lookup (VN aligns with membership UI ``buildSmsPhoneNumber``)."""
    d = normalize_digits(phone)
    if not d:
        return ""
    iso = (country_code or "").strip().upper()
    if iso == "VN":
        if d.startswith("0") and len(d) >= 2:
            return "84" + d[1:]
        if len(d) == 9 and not d.startswith("0"):
            return "84" + d
        if d.startswith("84") and len(d) >= 10:
            return d
    return d


def _load_raw_scenarios() -> list[dict[str, Any]]:
    try:
        with _MOCK_SCENARIOS_PATH.open(encoding="utf-8") as f:
            data: object = json.load(f)
    except OSError as err:
        logger.exception(
            "mock_scenarios_load_failed",
            extra={"path": str(_MOCK_SCENARIOS_PATH), "error": str(err)},
        )
        raise
    except json.JSONDecodeError as err:
        logger.exception(
            "mock_scenarios_json_invalid",
            extra={"path": str(_MOCK_SCENARIOS_PATH), "error": str(err)},
        )
        raise
    if not isinstance(data, dict):
        raise TypeError("mock_scenarios.json root must be an object")
    raw = data.get("scenarios")
    if not isinstance(raw, list):
        raise TypeError("mock_scenarios.json must contain a list 'scenarios'")
    out: list[dict[str, Any]] = []
    for item in raw:
        if isinstance(item, dict):
            out.append(item)
        else:
            logger.warning("mock_scenarios_skip_non_object", extra={"item": repr(item)})
    return out


def _build_index(scenarios: list[dict[str, Any]]) -> dict[str, dict[str, Any]]:
    index: dict[str, dict[str, Any]] = {}
    for scenario in scenarios:
        cc = scenario.get("country_code")
        pn = scenario.get("phone_number")
        cc_s = cc if isinstance(cc, str) else None
        pn_s = pn if isinstance(pn, str) else None
        key = canonical_mock_msisdn_digits(cc_s, pn_s)
        if not key:
            logger.warning(
                "mock_scenarios_skip_unindexed",
                extra={"country_code": cc_s, "phone_number": pn_s},
            )
            continue
        index[key] = scenario
    return index


_SCENARIOS: list[dict[str, Any]] = _load_raw_scenarios()
_MOCK_SCENARIO_BY_DIGITS: dict[str, dict[str, Any]] = _build_index(_SCENARIOS)


def lookup_mock_scenario(phone: str | None, country_code: str | None) -> dict[str, Any] | None:
    """Return the scenario dict for *phone* / *country_code*, or ``None`` if unknown."""
    key = canonical_mock_msisdn_digits(country_code, phone)
    if not key:
        return None
    return _MOCK_SCENARIO_BY_DIGITS.get(key)


def get_mock_outcome(phone: str | None, country_code: str | None) -> str | None:
    """Return outcome string from JSON (``send_success``, ``send_failed``, …) or ``None``."""
    scenario = lookup_mock_scenario(phone, country_code)
    if scenario is None:
        return None
    outcome = scenario.get("outcome")
    if isinstance(outcome, str) and outcome.strip():
        return outcome.strip()
    return None


def get_mock_actual_cost(phone: str | None, country_code: str | None) -> float | None:
    """Return ``actual_cost`` from the matching scenario, or ``None``."""
    scenario = lookup_mock_scenario(phone, country_code)
    if scenario is None:
        return None
    raw = scenario.get("actual_cost")
    if raw is None:
        return None
    if isinstance(raw, bool):
        return None
    if isinstance(raw, (int, float)):
        return float(raw)
    return None
