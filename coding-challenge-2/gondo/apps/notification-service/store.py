"""In-memory notification store (simulation-first)."""

from __future__ import annotations

from models import Notification, TransitionEvent

notifications: dict[str, Notification] = {}
notifications_by_message_id: dict[str, str] = {}
transition_events: dict[str, list[TransitionEvent]] = {}


def clear() -> None:
    """Reset all in-memory state (for tests)."""
    notifications.clear()
    notifications_by_message_id.clear()
    transition_events.clear()


def create_notification(notification: Notification) -> None:
    """Insert a new notification; caller must ensure IDs are unique."""
    notifications[notification.notification_id] = notification
    notifications_by_message_id[notification.message_id] = notification.notification_id


def get_notification(notification_id: str) -> Notification | None:
    return notifications.get(notification_id)


def find_by_message_id(message_id: str) -> Notification | None:
    nid = notifications_by_message_id.get(message_id)
    if nid is None:
        return None
    return notifications.get(nid)


def list_notifications() -> list[Notification]:
    return list(notifications.values())


def update_notification(notification: Notification) -> None:
    notifications[notification.notification_id] = notification
    notifications_by_message_id[notification.message_id] = notification.notification_id


def add_transition_event(event: TransitionEvent) -> None:
    bucket = transition_events.setdefault(event.notification_id, [])
    bucket.append(event)


def get_transition_events(notification_id: str) -> list[TransitionEvent]:
    return list(transition_events.get(notification_id, []))
