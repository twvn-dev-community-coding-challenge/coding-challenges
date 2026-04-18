# OpenAPI catalog (`docs/openapi/`)

This folder is the **published contract surface** for HTTP APIs whose JSON specs are checked into the repo.

| Artifact | Service | Purpose |
|----------|---------|---------|
| [notification-service.openapi.json](notification-service.openapi.json) | **notification-service** (platform SMS / User Story 2) | Import into Postman, API gateways, or code generators. |

## Regenerate

```bash
# from gondo/
yarn nx run notification-service:generate-openapi
```

Writes the file above and `apps/notification-service/openapi.json` (identical schema).

## Drift check (CI / pre-PR)

```bash
yarn verify-openapi
# or
yarn nx run notification-service:verify-openapi
```

Fails if the committed JSON does not match the live FastAPI `app.openapi()` output. The same check runs in **pytest** as `test_committed_openapi_json_matches_live_app`.

## TypeScript types (optional client)

```bash
yarn nx run ts-core:generate-openapi-types
```

Outputs `libs/ts-core/src/notification-api/openapi.generated.ts` (from `openapi-typescript`). Re-run after any contract change; the file is safe to commit for consumers that want `paths` / `components` types.

See [platform-sms-openapi.md](../platform-sms-openapi.md) for the product narrative.
