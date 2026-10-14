"""No-op placeholder — restores Alembic lineage when revision 0003 was stamped without a script.

Revision ID: 0003
Revises: 0002
Create Date: 2026-04-20
"""

from __future__ import annotations

from typing import Sequence, Union

revision: str = "0003"
down_revision: Union[str, None] = "0002"
branch_labels: Union[str, Sequence[str], None] = None
depends_on: Union[str, Sequence[str], None] = None


def upgrade() -> None:
    """Schema already complete at 0002; revision exists so ``alembic upgrade head`` can resolve."""
    pass


def downgrade() -> None:
    pass
