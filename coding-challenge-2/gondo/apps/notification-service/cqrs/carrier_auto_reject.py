"""Simulated MNO rejection for specific Vietnam test MSISDNs.

Applied **after** ``Queue`` → ``Send-to-carrier`` (carrier handoff): **Send-to-carrier** →
**Carrier-rejected**, as if the downstream MNO refused the MT toward the handset.

Maps national ``090909094`` and ``+8490909094`` to canonical digits ``8490909094``.
"""

from __future__ import annotations

_CARRIER_AUTO_REJECT_VN_MSISDN: frozenset[str] = frozenset({"8490909094"})


def vn_msisdn_digits(phone_number: str) -> str:
    """Canonical digit string for VN (aligned with membership UI / dev mock helpers)."""
    d = "".join(c for c in (phone_number or "").strip() if c.isdigit())
    if not d:
        return ""
    if d.startswith("0") and len(d) >= 2:
        return "84" + d[1:]
    if len(d) == 9 and not d.startswith("0"):
        return "84" + d
    if d.startswith("84") and len(d) >= 10:
        return d
    return ""


def should_auto_carrier_reject(country_code: str | None, phone_number: str) -> bool:
    """True when the carrier layer simulates MNO rejection for this recipient."""
    if (country_code or "").strip().upper() != "VN":
        return False
    return vn_msisdn_digits(phone_number) in _CARRIER_AUTO_REJECT_VN_MSISDN
