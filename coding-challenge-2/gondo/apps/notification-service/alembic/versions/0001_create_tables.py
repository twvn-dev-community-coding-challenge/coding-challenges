"""Create notification_service tables.

Revision ID: 0001
Revises: None
Create Date: 2026-04-12
"""

from __future__ import annotations

from typing import Sequence, Union

revision: str = "0001"
down_revision: Union[str, None] = None
branch_labels: Union[str, Sequence[str], None] = None
depends_on: Union[str, Sequence[str], None] = None


def upgrade() -> None:
    # notification_service no longer owns any tables.
    # carrier_prefixes has moved to provider_service schema.
    pass


def downgrade() -> None:
    pass
