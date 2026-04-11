"""PII redaction utilities for request/response payload logging."""

from __future__ import annotations

import re
from dataclasses import dataclass, field
from typing import TYPE_CHECKING, Literal

if TYPE_CHECKING:
    from google.protobuf.message import Message

StrategyName = Literal["mask_phone", "mask_email", "redact", "partial"]

_EMAIL_RE = re.compile(
    r"^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$",
)


@dataclass(frozen=True, slots=True)
class RedactionRule:
    """Associates a field name (exact or ``re:`` regex) with a redaction strategy."""

    field_pattern: str
    strategy: StrategyName


def mask_phone(value: object) -> str:
    """Mask a phone string: keep an optional ``+`` prefix (3 chars), mask the middle, keep last 4 digits."""
    if not isinstance(value, str):
        return "[REDACTED]"
    if len(value) < 7:
        return "[REDACTED]"
    if value.startswith("+"):
        prefix = value[:3]
        last4 = value[-4:]
        stars = "*" * (len(value) - 7)
        return f"{prefix}{stars}{last4}"
    prefix = value[:3]
    last4 = value[-4:]
    stars = "*" * (len(value) - 7)
    return f"{prefix}{stars}{last4}"


def mask_email(value: object) -> str:
    """Mask a valid email local part and domain for safe display."""
    if not isinstance(value, str):
        return "[REDACTED]"
    if "@" not in value:
        return "[REDACTED]"
    local, domain = value.split("@", 1)
    if not local or not domain:
        return "[REDACTED]"
    if not _EMAIL_RE.fullmatch(value):
        return "[REDACTED]"
    dot = domain.rfind(".")
    if dot <= 0:
        return "[REDACTED]"
    base = domain[:dot]
    tld = domain[dot:]
    if not base:
        return "[REDACTED]"
    masked_domain = f"{base[0]}***{tld}"
    if len(local) == 1:
        return f"{local}***@{masked_domain}"
    masked_local = f"{local[0]}***{local[-1]}"
    return f"{masked_local}@{masked_domain}"


def redact_full(_value: object) -> str:
    """Replace any value with a fixed redaction token."""
    return "[REDACTED]"


def partial_mask(value: object) -> str:
    """Show first and last character with a fixed mask in between."""
    if not isinstance(value, str):
        return "[REDACTED]"
    if len(value) <= 2:
        return "[REDACTED]"
    return f"{value[0]}***{value[-1]}"


def redact_value(value: object, strategy: StrategyName) -> str:
    """Apply a redaction strategy to a single value."""
    if strategy == "mask_phone":
        return mask_phone(value)
    if strategy == "mask_email":
        return mask_email(value)
    if strategy == "redact":
        return redact_full(value)
    if strategy == "partial":
        return partial_mask(value)
    raise ValueError(f"unknown strategy: {strategy!r}")


def _apply_strategy(value: str, strategy: StrategyName) -> str:
    if strategy == "mask_phone":
        return mask_phone(value)
    if strategy == "mask_email":
        return mask_email(value)
    if strategy == "redact":
        return redact_full(value)
    if strategy == "partial":
        return partial_mask(value)
    raise ValueError(f"unknown strategy: {strategy!r}")


@dataclass
class RedactionPolicy:
    """Ordered redaction rules applied to nested dict and list structures."""

    rules: tuple[RedactionRule, ...] = field(default_factory=tuple)

    def _find_matching_rule(self, key: str) -> RedactionRule | None:
        for rule in self.rules:
            if rule.field_pattern.startswith("re:"):
                pattern = rule.field_pattern[3:]
                if re.fullmatch(pattern, key):
                    return rule
            elif key == rule.field_pattern:
                return rule
        return None

    def redact_dict(self, data: dict[str, object]) -> dict[str, object]:
        """Return a new dict with redaction rules applied; input is not mutated."""
        result: dict[str, object] = {}
        for key, value in data.items():
            rule = self._find_matching_rule(key)
            if rule is not None:
                if rule.strategy == "redact":
                    result[key] = "[REDACTED]"
                elif isinstance(value, str):
                    result[key] = _apply_strategy(value, rule.strategy)
                elif isinstance(value, dict):
                    result[key] = self.redact_dict(value)
                else:
                    result[key] = "[REDACTED]"
            elif isinstance(value, dict):
                result[key] = self.redact_dict(value)
            elif isinstance(value, list):
                result[key] = [self._redact_list_item(item) for item in value]
            else:
                result[key] = value
        return result

    def _redact_list_item(self, item: object) -> object:
        if isinstance(item, dict):
            return self.redact_dict(item)
        if isinstance(item, list):
            return [self._redact_list_item(el) for el in item]
        return item

    def redact_proto(self, msg: Message) -> dict[str, object]:
        """Convert a protobuf message to a dict and apply ``redact_dict``."""
        from google.protobuf.json_format import MessageToDict

        raw = MessageToDict(msg, preserving_proto_field_name=True)
        return self.redact_dict(raw)


DEFAULT_POLICY = RedactionPolicy(
    rules=(
        RedactionRule(field_pattern="phone_number", strategy="mask_phone"),
        RedactionRule(field_pattern="phone", strategy="mask_phone"),
        RedactionRule(field_pattern="mobile", strategy="mask_phone"),
        RedactionRule(field_pattern="email", strategy="mask_email"),
        RedactionRule(
            field_pattern="re:(?i)(password|secret|token|api_key|authorization)",
            strategy="redact",
        ),
    ),
)
