"""Alembic environment for notification-service."""
from __future__ import annotations

import os
from logging.config import fileConfig

from alembic import context
from sqlalchemy import create_engine, text

config = context.config

if config.config_file_name is not None:
    fileConfig(config.config_file_name)

url = os.environ.get("DATABASE_URL", config.get_main_option("sqlalchemy.url"))
SCHEMA = "notification_service"


def run_migrations_online() -> None:
    engine = create_engine(url)
    with engine.connect() as conn:
        conn.execute(text(f"CREATE SCHEMA IF NOT EXISTS {SCHEMA}"))
        conn.commit()
        context.configure(
            connection=conn,
            target_metadata=None,
            version_table_schema=SCHEMA,
            include_schemas=True,
        )
        with context.begin_transaction():
            context.execute(text(f"SET search_path TO {SCHEMA}, public"))
            context.run_migrations()
        conn.commit()


run_migrations_online()
