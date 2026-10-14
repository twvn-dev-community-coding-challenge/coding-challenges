# Routing seeds vs challenge brief

PostgreSQL seeds live in **`apps/provider-service/alembic/versions/0002_seed_data.py`**. Provider IDs map to **`providers.code`**: `prv_01` Twilio, `prv_02` Vonage, `prv_03` Infobip, `prv_04` AWS SNS, `prv_05` Telnyx, `prv_06` MessageBird, `prv_07` Sinch.

## Vietnam — User Story 3 (updated agreements)

Effective **`2026-04-01`** onward (`routing_rule_version = 2`):

| Carrier (DB) | Required provider (brief) | Seed `provider_id` | Match |
|----------------|---------------------------|--------------------|-------|
| VIETTEL | Vonage | `prv_02` | Yes |
| MOBIFONE | Infobip | `prv_03` | Yes |
| VINAPHONE | Twilio | `prv_01` | Yes |

Earlier window **`2026-01-01` → `2026-04-01`** uses **User Story 1** illustrative routing (version `1`), intentionally different from Story 3 — demonstrates **time-effective** rules.

## Philippines — User Story 3

Effective **`2026-04-01`** (`routing_rule_version = 2`):

| Carrier | Required provider | Seed `provider_id` | Match |
|---------|-------------------|----------------------|-------|
| GLOBE | MessageBird | `prv_06` | Yes |
| SMART | Sinch | `prv_07` | Yes |
| DITO | MessageBird | `prv_06` | Yes |

## Thailand & Singapore

Story 1 routing remains as seeded (AIS/DTAC/Singtel/StarHub → agreed providers); Story 3 only changed **VN** and added **PH** in the brief used for this checkpoint.

---

*If the challenge document updates again, adjust **`0002_seed_data.py`** (or a follow-up migration) and refresh this table.*
