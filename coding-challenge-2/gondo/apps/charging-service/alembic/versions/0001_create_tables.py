"""Create charging_service tables.

Revision ID: 0001
Revises: None
Create Date: 2026-04-12
"""
from __future__ import annotations

from typing import Sequence, Union

from alembic import op
import sqlalchemy as sa

revision: str = "0001"
down_revision: Union[str, None] = None
branch_labels: Union[str, Sequence[str], None] = None
depends_on: Union[str, Sequence[str], None] = None

SCHEMA = "charging_service"


def upgrade() -> None:
    op.execute(sa.text("CREATE EXTENSION IF NOT EXISTS pgcrypto"))

    op.create_table(
        "rates",
        sa.Column(
            "id",
            sa.Uuid(as_uuid=True),
            primary_key=True,
            server_default=sa.text("gen_random_uuid()"),
        ),
        sa.Column("country_code", sa.CHAR(2), nullable=False),
        sa.Column("carrier_id", sa.Integer(), nullable=False),
        sa.Column("provider_id", sa.Integer(), nullable=False),
        sa.Column("currency", sa.CHAR(3), nullable=False),
        sa.Column("price_per_sms", sa.Numeric(12, 6), nullable=False),
        sa.Column("effective_from", sa.DateTime(timezone=True), nullable=False),
        sa.Column("effective_to", sa.DateTime(timezone=True), nullable=True),
        sa.CheckConstraint(
            "effective_to IS NULL OR effective_to > effective_from",
            name="rates_effective_chk",
        ),
        schema=SCHEMA,
    )
    op.create_index(
        "ix_rates_lookup",
        "rates",
        ["country_code", "carrier_id", "provider_id", "effective_from"],
        schema=SCHEMA,
    )


def downgrade() -> None:
    op.drop_index("ix_rates_lookup", table_name="rates", schema=SCHEMA)
    op.drop_table("rates", schema=SCHEMA)
