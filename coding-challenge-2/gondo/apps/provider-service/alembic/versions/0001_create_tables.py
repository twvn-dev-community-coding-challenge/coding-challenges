"""Create provider_service tables (PostgreSQL schema refactor).

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

SCHEMA = "provider_service"


def upgrade() -> None:
    op.execute(sa.text("CREATE EXTENSION IF NOT EXISTS pgcrypto"))

    op.create_table(
        "providers",
        sa.Column("id", sa.Integer(), primary_key=True, autoincrement=True),
        sa.Column("code", sa.String(32), nullable=False, unique=True),
        sa.Column("name", sa.String(128), nullable=False, unique=True),
        sa.Column(
            "created_at",
            sa.DateTime(timezone=True),
            nullable=False,
            server_default=sa.func.now(),
        ),
        schema=SCHEMA,
    )
    op.create_table(
        "carriers",
        sa.Column("id", sa.Integer(), primary_key=True, autoincrement=True),
        sa.Column("code", sa.String(64), nullable=False, unique=True),
        sa.Column("display_name", sa.String(128), nullable=False),
        sa.Column("country_code", sa.CHAR(2), nullable=False),
        sa.Column("status", sa.String(16), nullable=False),
        sa.CheckConstraint("status IN ('active', 'inactive')"),
        sa.Column(
            "created_at",
            sa.DateTime(timezone=True),
            nullable=False,
            server_default=sa.func.now(),
        ),
        schema=SCHEMA,
    )
    op.create_index(
        "ix_carriers_country",
        "carriers",
        ["country_code"],
        schema=SCHEMA,
    )
    op.create_table(
        "routing_rules",
        sa.Column(
            "id",
            sa.Uuid(as_uuid=True),
            primary_key=True,
            server_default=sa.text("gen_random_uuid()"),
        ),
        sa.Column("country_code", sa.CHAR(2), nullable=False),
        sa.Column(
            "carrier_id",
            sa.Integer(),
            sa.ForeignKey(f"{SCHEMA}.carriers.id"),
            nullable=False,
        ),
        sa.Column(
            "provider_id",
            sa.Integer(),
            sa.ForeignKey(f"{SCHEMA}.providers.id"),
            nullable=False,
        ),
        sa.Column("priority", sa.Integer(), nullable=False, server_default="100"),
        sa.Column("routing_rule_version", sa.Integer(), nullable=False),
        sa.Column("effective_from", sa.DateTime(timezone=True), nullable=False),
        sa.Column("effective_to", sa.DateTime(timezone=True), nullable=True),
        sa.CheckConstraint(
            "effective_to IS NULL OR effective_to > effective_from",
            name="routing_effective_chk",
        ),
        schema=SCHEMA,
    )
    op.create_index(
        "ix_routing_country_carrier_time",
        "routing_rules",
        ["country_code", "carrier_id", "effective_from"],
        schema=SCHEMA,
    )
    op.create_index(
        "ix_routing_active",
        "routing_rules",
        ["country_code", "carrier_id"],
        schema=SCHEMA,
        postgresql_where=sa.text("effective_to IS NULL"),
    )
    op.create_table(
        "carrier_prefixes",
        sa.Column(
            "id",
            sa.Uuid(as_uuid=True),
            primary_key=True,
            server_default=sa.text("gen_random_uuid()"),
        ),
        sa.Column("country_calling_code", sa.String(8), nullable=False),
        sa.Column("national_destination", sa.String(32), nullable=False),
        sa.Column(
            "carrier_id",
            sa.Integer(),
            sa.ForeignKey(f"{SCHEMA}.carriers.id"),
            nullable=False,
        ),
        sa.Column("match_priority", sa.Integer(), nullable=False),
        schema=SCHEMA,
    )
    op.execute(
        sa.text(
            f"CREATE INDEX ix_prefix_priority ON {SCHEMA}.carrier_prefixes "
            "(match_priority DESC)"
        )
    )


def downgrade() -> None:
    op.execute(sa.text(f"DROP INDEX IF EXISTS {SCHEMA}.ix_prefix_priority"))
    op.drop_table("carrier_prefixes", schema=SCHEMA)
    op.drop_index("ix_routing_active", table_name="routing_rules", schema=SCHEMA)
    op.drop_index(
        "ix_routing_country_carrier_time",
        table_name="routing_rules",
        schema=SCHEMA,
    )
    op.drop_table("routing_rules", schema=SCHEMA)
    op.drop_index("ix_carriers_country", table_name="carriers", schema=SCHEMA)
    op.drop_table("carriers", schema=SCHEMA)
    op.drop_table("providers", schema=SCHEMA)
