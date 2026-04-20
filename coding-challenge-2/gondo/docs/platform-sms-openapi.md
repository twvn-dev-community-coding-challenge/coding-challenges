# Platform SMS API — OpenAPI export (User Story 2)

**User Story 2** asks for **cross-domain reuse** of SMS capability: other teams should integrate against a **clear, versionable HTTP contract**, not only by reading Python code.

## Contract artifact

The **notification-service** HTTP API is the platform entry point for creating and driving notifications (including SMS). The machine-readable contract is exported as **OpenAPI 3** JSON:

| Location | Purpose |
|----------|---------|
| `docs/openapi/notification-service.openapi.json` | **Canonical repo copy** for reviewers, governance, and “drop in Postman / gateway” workflows. |
| `apps/notification-service/openapi.json` | Same schema, co-located with the service (handy for diffs next to code). |

Both files are produced by the same generator; keep them in sync.

## Regenerate

From `gondo/` (with `.venv` activated and `libs` on `PYTHONPATH`, as for other Python tasks):

```bash
yarn nx run notification-service:generate-openapi
```

This runs `apps/notification-service/generate_openapi.py`, which introspects the FastAPI `app` and writes both paths above.

## Drift check

```bash
yarn verify-openapi
```

Compares committed JSON to the live app (also covered by `test_committed_openapi_json_matches_live_app` in **notification-service** tests). A **GitHub Actions** job at **`.github/workflows/openapi-verify.yml`** (repository root) runs the same check on push/PR.

## TypeScript types

For a generated **OpenAPI-typed** surface (paths, operations, components):

```bash
yarn nx run ts-core:generate-openapi-types
```

Output: `libs/ts-core/src/notification-api/openapi.generated.ts` (regenerate after contract changes). See [docs/openapi/README.md](openapi/README.md).

## Consumer story

1. Import `docs/openapi/notification-service.openapi.json` into an API client, mock server, or API catalog.
2. Point the client at the deployed **notification-service** base URL (e.g. `http://localhost:8001` locally, or the mesh URL in Docker: `http://notification-service:8001` from sibling containers).
3. For flows that issue a **server-side OTP**, ensure **otp-service** is reachable from notification-service (`OTP_SERVICE_BASE_URL`) and use `issue_server_otp` on create where applicable — see root README and `apps/otp-service/`.

## Scope note

This export describes **notification-service** only. **otp-service**, **provider-service** (gRPC-first), and **carrier-service** expose separate interfaces; multi-service integration is documented in the main README architecture table.
