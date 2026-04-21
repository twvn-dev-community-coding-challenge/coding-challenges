# Multi-stage image for Python FastAPI + optional gRPC (charging, provider, notification).
# Build: docker build --build-arg SERVICE_NAME=charging-service --build-arg PORT=8003 -t charging .

ARG SERVICE_NAME
ARG PORT

# -----------------------------------------------------------------------------
# Builder: install dependencies into a venv
# -----------------------------------------------------------------------------
FROM python:3.13-slim AS builder

ARG SERVICE_NAME
WORKDIR /build

RUN python -m venv /opt/venv
ENV PATH="/opt/venv/bin:$PATH"

COPY libs/py-core/requirements.txt /tmp/py-core-requirements.txt
COPY apps/${SERVICE_NAME}/requirements.txt /tmp/requirements.txt
RUN pip install --no-cache-dir -r /tmp/py-core-requirements.txt -r /tmp/requirements.txt

# -----------------------------------------------------------------------------
# Runtime: minimal app tree + non-root user + curl for healthchecks
# -----------------------------------------------------------------------------
FROM python:3.13-slim AS runtime

ARG SERVICE_NAME
ARG PORT

ENV PORT=${PORT}
ENV PYTHONPATH=/app/libs/grpc-contracts:/app/libs/py-core

RUN apt-get update \
    && apt-get install -y --no-install-recommends curl \
    && rm -rf /var/lib/apt/lists/*

COPY --from=builder /opt/venv /opt/venv
ENV PATH="/opt/venv/bin:$PATH"

WORKDIR /app

COPY libs/grpc-contracts/__init__.py /app/libs/grpc-contracts/__init__.py
COPY libs/grpc-contracts/generated/ /app/libs/grpc-contracts/generated/
COPY libs/py-core/ /app/libs/py-core/
COPY infra/ /app/infra/
COPY apps/${SERVICE_NAME}/ /app/apps/${SERVICE_NAME}/

WORKDIR /app/apps/${SERVICE_NAME}

RUN useradd --create-home --uid 1000 appuser \
    && chown -R appuser:appuser /app

USER appuser

EXPOSE ${PORT}

CMD ["/bin/sh", "-c", "exec python -m uvicorn main:app --host 0.0.0.0 --port ${PORT}"]
