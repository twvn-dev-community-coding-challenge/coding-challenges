#!/usr/bin/env python3
"""Merge `logs/*.log` NDJSON into one readable SMS pipeline document (JSON).

Groups:
- Primary: ``trace_id`` (when not ``-``) — follows UI → notification → gRPC → bus → carrier → notification subscriber.
- Secondary: attach orphan lines (provider/carrier) that share ``message_id`` + timestamp window.

Usage::
    cd gondo && python3 scripts/aggregate_sms_flow_logs.py
    python3 scripts/aggregate_sms_flow_logs.py --logs-dir logs --output logs/sms-flow-aggregate.log
"""

from __future__ import annotations

import argparse
import json
from collections import defaultdict
from collections.abc import Iterator
from datetime import datetime, timezone
from pathlib import Path
from typing import Any


INTERESTING_MESSAGE_NAMES: frozenset[str] = frozenset(
    {
        "notification_dispatch_flow",
        "notification_retry_flow",
        "notification_charging_estimate_step",
        "grpc_client_call_begin",
        "grpc_client_call_ok",
        "provider_grpc_request",
        "provider_grpc_response",
        "charging_grpc_request",
        "charging_grpc_response",
        "bus_publish_begin",
        "carrier_consume_begin",
        "carrier_dispatch_requested_parsed",
        "notification_consume_begin",
        "notification_dispatch_received_parsed",
        "carrier_dispatch_received_applied",
        "sms_dispatch_requested_published",
        "sms_dispatch_received_published",
        "sms_dispatch_outcome_published",
        "sms_dispatch_outcome_received",
        "provider_consume_begin",
        "http_request",
        "http_response",
        "grpc_server_request",
        "grpc_server_response",
        "grpc_client_request",
        "grpc_client_response",
    }
)

INTERESTING_LOGGER_PREFIXES: tuple[str, ...] = (
    "main",
    "grpc_client",
    "charging_client",
    "grpc_server",
    "cqrs.",
    "lifespan",
    "py_core.payload.",
)


def _parse_ts(raw: str) -> datetime | None:
    try:
        normalized = raw.replace("Z", "+00:00")
        dt = datetime.fromisoformat(normalized)
        if dt.tzinfo is None:
            dt = dt.replace(tzinfo=timezone.utc)
        return dt.astimezone(timezone.utc)
    except (TypeError, ValueError):
        return None


def _walk_log_files(logs_dir: Path) -> list[Path]:
    return sorted(logs_dir.glob("*.log"))


def _iter_records(path: Path) -> Iterator[dict[str, Any]]:
    with path.open(encoding="utf-8", errors="replace") as f:
        for line_no, line in enumerate(f, 1):
            line = line.strip()
            if not line:
                continue
            try:
                obj = json.loads(line)
            except json.JSONDecodeError:
                continue
            if isinstance(obj, dict):
                yield obj


def _is_interesting(rec: Any) -> bool:
    if not isinstance(rec, dict):
        return False
    msg = rec.get("message")
    if msg in INTERESTING_MESSAGE_NAMES:
        return True
    name = str(rec.get("name") or "")
    if name.startswith(INTERESTING_LOGGER_PREFIXES):
        return True
    if msg in ("grpc_server_request", "grpc_server_response", "grpc_client_request", "grpc_client_response"):
        return True
    return False


def _slim_record(rec: dict[str, Any]) -> dict[str, Any]:
    skip = frozenset({"exc_info"})
    out: dict[str, Any] = {}
    for k, v in rec.items():
        if k in skip:
            continue
        if k == "message" and isinstance(v, str) and len(v) > 500:
            out[k] = v[:500] + "…"
            continue
        out[k] = v
    return out


def _normalize_trace_id(rec: dict[str, Any]) -> str | None:
    tid = rec.get("trace_id")
    if tid is None or tid == "-" or tid == "":
        return None
    return str(tid)


def _extract_message_id(rec: dict[str, Any]) -> str | None:
    mid = rec.get("message_id")
    if isinstance(mid, str) and mid.strip():
        return mid.strip()
    body = rec.get("body")
    if isinstance(body, dict):
        pid = body.get("message_id")
        if isinstance(pid, str) and pid.strip():
            return pid.strip()
        pc = body.get("policy_context")
        if isinstance(pc, dict):
            pid = pc.get("message_id")
            if isinstance(pid, str) and pid.strip():
                return pid.strip()
    return None


def _is_sms_dispatch_pipeline_trace(flow_steps: list[dict[str, Any]]) -> bool:
    markers = frozenset({"notification_dispatch_flow", "notification_retry_flow"})
    return any(str(s.get("message")) in markers for s in flow_steps)


def collect_flows(logs_dir: Path, *, include_all_traces: bool) -> dict[str, Any]:
    source_files = [p.name for p in _walk_log_files(logs_dir)]
    by_trace: dict[str, list[dict[str, Any]]] = defaultdict(list)
    by_message_loose: dict[str, list[dict[str, Any]]] = defaultdict(list)

    for log_path in _walk_log_files(logs_dir):
        service_hint = log_path.stem.replace("-", "_")
        for rec in _iter_records(log_path):
            if not _is_interesting(rec):
                continue
            slim = _slim_record(dict(rec))
            slim["_log_file"] = log_path.name
            if "service" not in slim:
                slim["service"] = slim.get("service") or service_hint

            tid = _normalize_trace_id(rec)
            if tid:
                by_trace[tid].append(slim)
            mid = _extract_message_id(rec)
            if mid and not tid:
                by_message_loose[mid].append(slim)

    # Merge loose rows into traces when message_id matches a trace that has dispatch flow
    trace_message_ids: dict[str, set[str]] = defaultdict(set)
    for tid, rows in by_trace.items():
        for r in rows:
            mid = _extract_message_id(r)
            if mid:
                trace_message_ids[tid].add(mid)

    attached_loose: dict[str, list[str]] = defaultdict(list)
    for tid, mids in trace_message_ids.items():
        if not mids:
            continue
        seen_fp = {
            (r.get("timestamp"), r.get("message"), r.get("_log_file"), r.get("service"))
            for r in by_trace[tid]
        }
        for mid in mids:
            for loose in by_message_loose.get(mid, []):
                fp = (
                    loose.get("timestamp"),
                    loose.get("message"),
                    loose.get("_log_file"),
                    loose.get("service"),
                )
                if fp in seen_fp:
                    continue
                ts_loose = _parse_ts(str(loose.get("timestamp", "")))
                if ts_loose is None:
                    continue
                window_min: datetime | None = None
                window_max: datetime | None = None
                for r in by_trace[tid]:
                    if _extract_message_id(r) != mid:
                        continue
                    ts = _parse_ts(str(r.get("timestamp", "")))
                    if ts is None:
                        continue
                    window_min = ts if window_min is None or ts < window_min else window_min
                    window_max = ts if window_max is None or ts > window_max else window_max
                if window_min is None:
                    continue
                pad = 120.0  # seconds — cover async carrier → notification
                start = window_min.timestamp() - 5
                end = window_max.timestamp() + pad
                if start <= ts_loose.timestamp() <= end:
                    dup = dict(loose)
                    dup["_attached_via"] = "message_id_time_window"
                    by_trace[tid].append(dup)
                    seen_fp.add(fp)
                    attached_loose[tid].append(mid)

    # Sort each trace by timestamp
    def sort_key(r: dict[str, Any]) -> tuple[float, str]:
        ts = _parse_ts(str(r.get("timestamp", "")))
        return (ts.timestamp() if ts else 0.0, str(r.get("message", "")))

    flows: list[dict[str, Any]] = []
    for tid in sorted(by_trace.keys()):
        rows = sorted(by_trace[tid], key=sort_key)
        mids = trace_message_ids.get(tid, set())
        flows.append(
            {
                "trace_id": tid,
                "correlation": {
                    "message_ids": sorted(mids),
                    "notification_id": next(
                        (
                            r.get("notification_id")
                            for r in rows
                            if r.get("notification_id")
                        ),
                        None,
                    ),
                },
                "step_count": len(rows),
                "steps": rows,
            }
        )

    sms_dispatch_pipeline = [f for f in flows if _is_sms_dispatch_pipeline_trace(f["steps"])]

    doc: dict[str, Any] = {
        "generated_at": datetime.now(timezone.utc).isoformat(),
        "logs_directory": str(logs_dir.resolve()),
        "source_files": source_files,
        "summary": {
            "total_trace_groups": len(flows),
            "sms_dispatch_pipeline_trace_groups": len(sms_dispatch_pipeline),
        },
        "sms_dispatch_pipeline": sms_dispatch_pipeline,
        "notes": (
            "Primary array ``sms_dispatch_pipeline``: traces that include notification_dispatch_flow / "
            "notification_retry_flow (UI dispatch → provider → charging → bus → downstream). "
            "Rows from provider/carrier without trace_id may be merged by message_id + time window."
        ),
    }
    if include_all_traces:
        doc["all_trace_groups"] = flows
    return doc


def main() -> None:
    parser = argparse.ArgumentParser(description="Aggregate SMS pipeline logs into one JSON file.")
    parser.add_argument(
        "--logs-dir",
        type=Path,
        default=Path(__file__).resolve().parent.parent / "logs",
        help="Directory containing *.log NDJSON files",
    )
    parser.add_argument(
        "--output",
        type=Path,
        default=Path(__file__).resolve().parent.parent / "logs" / "sms-flow-aggregate.log",
        help="Output path (JSON content; default logs/sms-flow-aggregate.log)",
    )
    parser.add_argument(
        "--all-traces",
        action="store_true",
        help="Include ``all_trace_groups`` with every trace_id bucket (large/noisy).",
    )
    args = parser.parse_args()

    if not args.logs_dir.is_dir():
        raise SystemExit(f"Logs directory not found: {args.logs_dir}")

    doc = collect_flows(args.logs_dir, include_all_traces=args.all_traces)
    args.output.parent.mkdir(parents=True, exist_ok=True)
    args.output.write_text(json.dumps(doc, indent=2, ensure_ascii=False) + "\n", encoding="utf-8")
    n_pipe = doc["summary"]["sms_dispatch_pipeline_trace_groups"]
    n_tot = doc["summary"]["total_trace_groups"]
    print(f"Wrote {args.output} ({n_pipe} SMS dispatch traces; {n_tot} total trace groups)")


if __name__ == "__main__":
    main()
