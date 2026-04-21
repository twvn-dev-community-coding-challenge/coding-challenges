"""Seed providers, carriers, routing rules, and carrier prefixes.

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

SCHEMA = "provider_service"


def upgrade() -> None:
    op.execute(
        sa.text(
            f"""
            INSERT INTO {SCHEMA}.providers (id, code, name, created_at) VALUES
              (1, 'prv_01', 'Twilio',      '2026-01-01T00:00:00+00:00'::timestamptz),
              (2, 'prv_02', 'Vonage',      '2026-01-01T00:00:00+00:00'::timestamptz),
              (3, 'prv_03', 'Infobip',     '2026-01-01T00:00:00+00:00'::timestamptz),
              (4, 'prv_04', 'AWS SNS',     '2026-01-01T00:00:00+00:00'::timestamptz),
              (5, 'prv_05', 'Telnyx',      '2026-01-01T00:00:00+00:00'::timestamptz),
              (6, 'prv_06', 'MessageBird', '2026-01-01T00:00:00+00:00'::timestamptz),
              (7, 'prv_07', 'Sinch',       '2026-01-01T00:00:00+00:00'::timestamptz)
            """
        )
    )
    op.execute(
        sa.text(
            f"""
            SELECT setval(
              pg_get_serial_sequence('{SCHEMA}.providers', 'id'),
              (SELECT MAX(id) FROM {SCHEMA}.providers)
            )
            """
        )
    )

    op.execute(
        sa.text(
            f"""
            INSERT INTO {SCHEMA}.carriers (id, code, display_name, country_code, status, created_at) VALUES
              (1,  'VIETTEL',   'Viettel',   'VN', 'active', '2026-01-01T00:00:00+00:00'::timestamptz),
              (2,  'MOBIFONE',  'Mobifone',  'VN', 'active', '2026-01-01T00:00:00+00:00'::timestamptz),
              (3,  'VINAPHONE', 'Vinaphone', 'VN', 'active', '2026-01-01T00:00:00+00:00'::timestamptz),
              (4,  'AIS',       'AIS',       'TH', 'active', '2026-01-01T00:00:00+00:00'::timestamptz),
              (5,  'DTAC',      'DTAC',      'TH', 'active', '2026-01-01T00:00:00+00:00'::timestamptz),
              (6,  'SINGTEL',   'Singtel',   'SG', 'active', '2026-01-01T00:00:00+00:00'::timestamptz),
              (7,  'STARHUB',   'StarHub',   'SG', 'active', '2026-01-01T00:00:00+00:00'::timestamptz),
              (8,  'GLOBE',     'Globe',     'PH', 'active', '2026-01-01T00:00:00+00:00'::timestamptz),
              (9,  'SMART',     'Smart',     'PH', 'active', '2026-01-01T00:00:00+00:00'::timestamptz),
              (10, 'DITO',      'DITO',      'PH', 'active', '2026-01-01T00:00:00+00:00'::timestamptz)
            """
        )
    )
    op.execute(
        sa.text(
            f"""
            SELECT setval(
              pg_get_serial_sequence('{SCHEMA}.carriers', 'id'),
              (SELECT MAX(id) FROM {SCHEMA}.carriers)
            )
            """
        )
    )

    op.execute(
        sa.text(
            f"""
            INSERT INTO {SCHEMA}.routing_rules (
              id, country_code, carrier_id, provider_id, priority, routing_rule_version, effective_from, effective_to
            ) VALUES
              ('a0000001-0000-4000-8000-000000000001'::uuid, 'VN', 1, 1, 100, 1, '2026-01-01T00:00:00+00:00'::timestamptz, '2026-04-01T00:00:00+00:00'::timestamptz),
              ('a0000002-0000-4000-8000-000000000002'::uuid, 'VN', 2, 2, 100, 1, '2026-01-01T00:00:00+00:00'::timestamptz, '2026-04-01T00:00:00+00:00'::timestamptz),
              ('a0000003-0000-4000-8000-000000000003'::uuid, 'VN', 3, 2, 100, 1, '2026-01-01T00:00:00+00:00'::timestamptz, '2026-04-01T00:00:00+00:00'::timestamptz),
              ('a0000004-0000-4000-8000-000000000004'::uuid, 'TH', 4, 3, 100, 1, '2026-01-01T00:00:00+00:00'::timestamptz, NULL),
              ('a0000005-0000-4000-8000-000000000005'::uuid, 'TH', 5, 4, 100, 1, '2026-01-01T00:00:00+00:00'::timestamptz, NULL),
              ('a0000006-0000-4000-8000-000000000006'::uuid, 'SG', 6, 1, 100, 1, '2026-01-01T00:00:00+00:00'::timestamptz, NULL),
              ('a0000007-0000-4000-8000-000000000007'::uuid, 'SG', 7, 5, 100, 1, '2026-01-01T00:00:00+00:00'::timestamptz, NULL),
              ('a0000008-0000-4000-8000-000000000008'::uuid, 'VN', 1, 2, 100, 2, '2026-04-01T00:00:00+00:00'::timestamptz, NULL),
              ('a0000009-0000-4000-8000-000000000009'::uuid, 'VN', 2, 3, 100, 2, '2026-04-01T00:00:00+00:00'::timestamptz, NULL),
              ('a000000a-0000-4000-8000-00000000000a'::uuid, 'VN', 3, 1, 100, 2, '2026-04-01T00:00:00+00:00'::timestamptz, NULL),
              ('a000000b-0000-4000-8000-00000000000b'::uuid, 'PH', 8, 6, 100, 2, '2026-04-01T00:00:00+00:00'::timestamptz, NULL),
              ('a000000c-0000-4000-8000-00000000000c'::uuid, 'PH', 9, 7, 100, 2, '2026-04-01T00:00:00+00:00'::timestamptz, NULL),
              ('a000000d-0000-4000-8000-00000000000d'::uuid, 'PH', 10, 6, 100, 2, '2026-04-01T00:00:00+00:00'::timestamptz, NULL)
            """
        )
    )

    op.execute(
        sa.text(
            f"""
            INSERT INTO {SCHEMA}.carrier_prefixes (
              id, country_calling_code, national_destination, carrier_id, match_priority
            ) VALUES
              ('c0000001-0000-4000-8000-000000000001'::uuid, '84', '39', 1, 50),
              ('c0000002-0000-4000-8000-000000000002'::uuid, '84', '38', 1, 50),
              ('c0000003-0000-4000-8000-000000000003'::uuid, '84', '37', 1, 50),
              ('c0000004-0000-4000-8000-000000000004'::uuid, '84', '36', 1, 50),
              ('c0000005-0000-4000-8000-000000000005'::uuid, '84', '35', 1, 50),
              ('c0000006-0000-4000-8000-000000000006'::uuid, '84', '34', 1, 50),
              ('c0000007-0000-4000-8000-000000000007'::uuid, '84', '33', 1, 50),
              ('c0000008-0000-4000-8000-000000000008'::uuid, '84', '32', 1, 50),
              ('c0000009-0000-4000-8000-000000000009'::uuid, '84', '96', 1, 50),
              ('c000000a-0000-4000-8000-00000000000a'::uuid, '84', '97', 1, 50),
              ('c000000b-0000-4000-8000-00000000000b'::uuid, '84', '98', 1, 50),
              ('c000000c-0000-4000-8000-00000000000c'::uuid, '84', '90', 2, 50),
              ('c000000d-0000-4000-8000-00000000000d'::uuid, '84', '93', 2, 50),
              ('c000000e-0000-4000-8000-00000000000e'::uuid, '84', '89', 2, 50),
              ('c000000f-0000-4000-8000-00000000000f'::uuid, '84', '70', 2, 50),
              ('c0000010-0000-4000-8000-000000000010'::uuid, '84', '79', 2, 50),
              ('c0000011-0000-4000-8000-000000000011'::uuid, '84', '91', 3, 50),
              ('c0000012-0000-4000-8000-000000000012'::uuid, '84', '94', 3, 50),
              ('c0000013-0000-4000-8000-000000000013'::uuid, '84', '88', 3, 50),
              ('c0000014-0000-4000-8000-000000000014'::uuid, '84', '83', 3, 50),
              ('c0000015-0000-4000-8000-000000000015'::uuid, '84', '84', 3, 50),
              ('c0000016-0000-4000-8000-000000000016'::uuid, '66', '81', 4, 40),
              ('c0000017-0000-4000-8000-000000000017'::uuid, '66', '82', 5, 40),
              ('c0000018-0000-4000-8000-000000000018'::uuid, '65', '8', 6, 30),
              ('c0000019-0000-4000-8000-000000000019'::uuid, '65', '9', 7, 30),
              ('c000001a-0000-4000-8000-00000000001a'::uuid, '63', '917', 8, 60),
              ('c000001b-0000-4000-8000-00000000001b'::uuid, '63', '905', 8, 60),
              ('c000001c-0000-4000-8000-00000000001c'::uuid, '63', '918', 9, 60),
              ('c000001d-0000-4000-8000-00000000001d'::uuid, '63', '999', 10, 60)
            """
        )
    )


def downgrade() -> None:
    op.execute(sa.text(f"DELETE FROM {SCHEMA}.carrier_prefixes"))
    op.execute(sa.text(f"DELETE FROM {SCHEMA}.routing_rules"))
    op.execute(sa.text(f"DELETE FROM {SCHEMA}.carriers"))
    op.execute(sa.text(f"DELETE FROM {SCHEMA}.providers"))
