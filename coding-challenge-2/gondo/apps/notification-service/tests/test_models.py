"""Tests for domain models and state machine."""

from __future__ import annotations

from models import VALID_TRANSITIONS, is_valid_transition


def test_valid_transition_new_to_send_to_provider() -> None:
    assert is_valid_transition("New", "Send-to-provider") is True


def test_invalid_transition_new_to_queue() -> None:
    assert is_valid_transition("New", "Queue") is False


def test_terminal_state_send_success_has_no_transitions() -> None:
    assert "Send-success" not in VALID_TRANSITIONS
    assert is_valid_transition("Send-success", "Send-failed") is False


def test_send_to_carrier_to_carrier_rejected_allowed() -> None:
    assert is_valid_transition("Send-to-carrier", "Carrier-rejected") is True
