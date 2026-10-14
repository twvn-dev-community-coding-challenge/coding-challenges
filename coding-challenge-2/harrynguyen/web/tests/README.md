# Playwright Registration E2E

This test suite maps to the registration matrix IDs `REG-001` to `REG-015`.

## Run

From `web/`:

- Install deps: `npm install`
- Run tests: `npm run test:e2e`

## Required API Base URL

- `PLAYWRIGHT_API_BASE_URL` (default: `http://127.0.0.1:8080`)

## Endpoint Overrides

- `PW_REGISTER_START_ENDPOINT` (default: `/api/register`)
- `PW_REGISTER_VERIFY_ENDPOINT` (default: `/api/verify`)
- `PW_VERIFY_ID_FIELD` (default: `username`, set to `registration_id` for new contract)

## Optional Test Hooks (for advanced cases)

- `PW_TEST_USER_LOOKUP_ENDPOINT` (used by REG-014)
- `PW_TEST_AUDIT_ENDPOINT` (used by REG-015)
- `PW_TEST_SUPPORTS_PROVIDER_FAILURE_SIM` (enables REG-005 and REG-006)
- `PW_TEST_EXPIRED_OTP_CODE` (enables REG-008)
- `PW_TEST_SUPPORTS_RATE_LIMIT_RESET` (enables REG-010, REG-011, REG-012)

If optional hooks are not set, related test cases are skipped.
