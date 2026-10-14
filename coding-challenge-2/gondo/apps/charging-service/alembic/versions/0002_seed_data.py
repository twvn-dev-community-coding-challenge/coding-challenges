"""Seed rates.

Revision ID: 0002
Revises: 0001
Create Date: 2026-04-12
"""
from __future__ import annotations

from typing import Sequence, Union

from alembic import op
import sqlalchemy as sa

revision: str = "0002"
down_revision: Union[str, None] = "0001"
branch_labels: Union[str, Sequence[str], None] = None
depends_on: Union[str, Sequence[str], None] = None

SCHEMA = "charging_service"


def upgrade() -> None:
    op.execute(
        sa.text(
            f"""
            INSERT INTO {SCHEMA}.rates (
              id, country_code, carrier_id, provider_id, currency, price_per_sms, effective_from, effective_to
            ) VALUES
              ('b0000001-0000-4000-8000-000000000001'::uuid, 'VN', 1, 1, 'USD', 0.015000, '2026-01-01T00:00:00+00:00', '2026-04-01T00:00:00+00:00'),
              ('b0000002-0000-4000-8000-000000000002'::uuid, 'VN', 2, 2, 'USD', 0.018000, '2026-01-01T00:00:00+00:00', '2026-04-01T00:00:00+00:00'),
              ('b0000003-0000-4000-8000-000000000003'::uuid, 'VN', 3, 2, 'USD', 0.012000, '2026-01-01T00:00:00+00:00', '2026-04-01T00:00:00+00:00'),
              ('b0000004-0000-4000-8000-000000000004'::uuid, 'VN', 1, 2, 'USD', 0.030000, '2026-04-01T00:00:00+00:00', NULL),
              ('b0000005-0000-4000-8000-000000000005'::uuid, 'VN', 2, 3, 'USD', 0.019500, '2026-04-01T00:00:00+00:00', NULL),
              ('b0000006-0000-4000-8000-000000000006'::uuid, 'VN', 3, 1, 'USD', 0.020500, '2026-04-01T00:00:00+00:00', NULL),
              ('b0000007-0000-4000-8000-000000000007'::uuid, 'TH', 4, 3, 'USD', 0.022000, '2026-01-01T00:00:00+00:00', NULL),
              ('b0000008-0000-4000-8000-000000000008'::uuid, 'TH', 5, 4, 'USD', 0.021000, '2026-01-01T00:00:00+00:00', NULL),
              ('b0000009-0000-4000-8000-000000000009'::uuid, 'SG', 6, 1, 'USD', 0.009500, '2026-01-01T00:00:00+00:00', NULL),
              ('b000000a-0000-4000-8000-00000000000a'::uuid, 'SG', 7, 5, 'USD', 0.010200, '2026-01-01T00:00:00+00:00', NULL),
              ('b000000b-0000-4000-8000-00000000000b'::uuid, 'PH', 8, 6, 'USD', 0.018500, '2026-04-01T00:00:00+00:00', NULL),
              ('b000000c-0000-4000-8000-00000000000c'::uuid, 'PH', 9, 7, 'USD', 0.017800, '2026-04-01T00:00:00+00:00', NULL),
              ('b000000d-0000-4000-8000-00000000000d'::uuid, 'PH', 10, 6, 'USD', 0.019000, '2026-04-01T00:00:00+00:00', NULL);
            """
        )
    )


def downgrade() -> None:
    op.execute(f"DELETE FROM {SCHEMA}.rates")
