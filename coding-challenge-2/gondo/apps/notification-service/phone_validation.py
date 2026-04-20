"""Phone number normalization (E.164) and country-specific validation for SMS."""

from __future__ import annotations

import re
from typing import Final

COUNTRY_CALLING_CODES: Final[dict[str, str]] = {
    "VN": "84",
    "TH": "66",
    "SG": "65",
    "PH": "63",
}

# Two-digit carrier prefix after national trunk 0 (national = 0 + 9 subscriber digits).
_VN_TWO_DIGIT_PREFIXES: Final[frozenset[str]] = frozenset(
    [f"{n:02d}" for n in range(32, 40)]
    + [
        "70",
        "79",
        "83",
        "84",
        "88",
        "89",
        "90",
        "91",
        "93",
        "94",
        "96",
        "97",
        "98",
        "99",
    ]
)

_TH_TWO_DIGIT_PREFIXES: Final[frozenset[str]] = frozenset({"81", "82"})

_PH_THREE_DIGIT_PREFIXES: Final[frozenset[str]] = frozenset(
    {"917", "905", "918", "999"},
)


def _digits_only(phone_number: str) -> str:
    return re.sub(r"\D", "", phone_number)


def _e164(calling_code: str, subscriber_digits: str) -> str:
    return f"+{calling_code}{subscriber_digits}"


def normalize_phone_number(country_code: str, phone_number: str) -> str:
    """Normalize phone_number to E.164 format.

    - Strips non-digit chars (except leading +)
    - For VN: handles national (0XXXXXXXXX) and E.164 (+84...)
    - For TH: handles national (0XXXXXXXXX) and E.164 (+66...)
    - For SG: handles national (8 digits) and E.164 (+65...)
    - For PH: handles national (0XXXXXXXXXX) and E.164 (+63...)
    - Unsupported countries: digits only with a leading + (no country-specific rules)
    """
    cc = country_code.strip().upper()
    digits = _digits_only(phone_number)
    if not digits:
        return "+"

    if cc not in COUNTRY_CALLING_CODES:
        return f"+{digits}"

    calling = COUNTRY_CALLING_CODES[cc]

    if cc == "VN":
        if len(digits) == 11 and digits.startswith(calling):
            return _e164(calling, digits[len(calling) :])
        if len(digits) == 10 and digits.startswith("0"):
            return _e164(calling, digits[1:])
        return f"+{digits}"

    if cc == "TH":
        if len(digits) == 11 and digits.startswith(calling):
            return _e164(calling, digits[len(calling) :])
        if len(digits) == 10 and digits.startswith("0"):
            return _e164(calling, digits[1:])
        if len(digits) == 9 and not digits.startswith(calling):
            return _e164(calling, digits)
        return f"+{digits}"

    if cc == "SG":
        if len(digits) == 10 and digits.startswith(calling):
            return _e164(calling, digits[len(calling) :])
        if len(digits) == 8 and digits[0] in ("8", "9"):
            return _e164(calling, digits)
        return f"+{digits}"

    if cc == "PH":
        if len(digits) == 12 and digits.startswith(calling):
            return _e164(calling, digits[len(calling) :])
        if len(digits) == 11 and digits.startswith("0"):
            return _e164(calling, digits[1:])
        if len(digits) == 10 and not digits.startswith(calling):
            return _e164(calling, digits)
        return f"+{digits}"

    return f"+{digits}"


def validate_phone_number(country_code: str, phone_number: str) -> str | None:
    """Validate the phone number for the given country.

    Expects ``phone_number`` in E.164 form for supported countries (output of
    :func:`normalize_phone_number`). Returns ``None`` if valid, or a short
    error message if invalid.

    Unsupported country codes are not validated (returns ``None``).
    """
    cc = country_code.strip().upper()
    if cc not in COUNTRY_CALLING_CODES:
        return None

    calling = COUNTRY_CALLING_CODES[cc]
    prefix = f"+{calling}"
    if not phone_number.startswith(prefix):
        return (
            f"Invalid phone number for {cc}: expected E.164 with prefix {prefix} "
            f"(e.g. national 0… or +{calling}…)."
        )

    subscriber = phone_number[len(prefix) :]
    if not subscriber.isdigit():
        return f"Invalid phone number for {cc}: only digits are allowed after {prefix}."

    if cc == "VN":
        if len(subscriber) != 9:
            return (
                "Invalid Vietnam phone number: after +84 there must be exactly 9 digits "
                f"(national 0 + 9 digits); got {len(subscriber)}."
            )
        two = subscriber[:2]
        if two not in _VN_TWO_DIGIT_PREFIXES:
            allowed = ", ".join(sorted(_VN_TWO_DIGIT_PREFIXES))
            return (
                "Invalid Vietnam phone number: the digit pair after the national leading 0 "
                f"must be a known mobile prefix (e.g. 32–39, 70, 79, 83–84, 88–91, 93–94, 96–98, 99); "
                f"got “{two}”. Allowed: {allowed}."
            )
        return None

    if cc == "TH":
        if len(subscriber) != 9:
            return (
                "Invalid Thailand phone number: after +66 there must be exactly 9 digits "
                f"(national 0 + 9 digits); got {len(subscriber)}."
            )
        two = subscriber[:2]
        if two not in _TH_TWO_DIGIT_PREFIXES:
            return (
                "Invalid Thailand phone number: after the national leading 0 the number "
                f"must start with carrier prefix 81 or 82; got “{two}”."
            )
        return None

    if cc == "SG":
        if len(subscriber) != 8:
            return (
                "Invalid Singapore phone number: after +65 there must be exactly 8 digits; "
                f"got {len(subscriber)}."
            )
        if subscriber[0] not in ("8", "9"):
            return (
                "Invalid Singapore phone number: the first digit must be 8 or 9 "
                f"(national 8-digit mobile); got “{subscriber[0]}”."
            )
        return None

    if cc == "PH":
        if len(subscriber) != 10:
            return (
                "Invalid Philippines phone number: after +63 there must be exactly 10 digits "
                f"(national 0 + 10 digits); got {len(subscriber)}."
            )
        three = subscriber[:3]
        if three not in _PH_THREE_DIGIT_PREFIXES:
            return (
                "Invalid Philippines phone number: after the national leading 0 the number "
                f"must start with carrier prefix 905, 917, 918, or 999; got “{three}”."
            )
        return None

    return None
