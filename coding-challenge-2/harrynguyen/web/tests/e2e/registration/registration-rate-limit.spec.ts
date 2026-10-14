import { randomInt } from "node:crypto";
import { test, expect } from "../../fixtures/registration.fixture";
import type { StartRegistrationResponse } from "../../fixtures/registration.fixture";
import { makeRegistrationPayload } from "../../helpers/test-data";
import { resolveVerifyIdentifier } from "../../helpers/otp-helper";

test.describe("Registration throttling and observability", () => {
  test("REG-010: maximum 10 SMS per hour per phone number is enforced", async ({ request }) => {
    const key =
      process.env.PW_SMS_INTERNAL_API_KEY || process.env.SMS_INTERNAL_API_KEY;
    test.skip(
      !key,
      "Set PW_SMS_INTERNAL_API_KEY or SMS_INTERNAL_API_KEY to match the API server (playwright.config loads repo-root .env)",
    );

    const phone = `63917${String(randomInt(1_000_000, 9_999_999))}`;
    let lastStatus = 0;
    for (let i = 0; i < 11; i += 1) {
      const response = await request.post("/api/sms/send", {
        headers: { Authorization: `Bearer ${key}` },
        data: { country: "PH", phone_number: phone, content: `rate-limit probe ${i}` },
      });
      lastStatus = response.status();
    }

    expect(lastStatus).toBe(429);
  });

  test("REG-011: intensive restart by IP is throttled", async ({ registrationApi }) => {
    let blocked = false;
    for (let i = 0; i < 7; i += 1) {
      const payload = makeRegistrationPayload();
      const response = await registrationApi.startRegistration(payload, {
        "x-forwarded-for": "10.10.10.10",
      });
      if (response.status() === 429) {
        blocked = true;
        break;
      }
    }

    expect(blocked).toBe(true);
  });

  test("REG-012: intensive restart by device fingerprint is throttled", async ({ registrationApi }) => {
    let blocked = false;
    for (let i = 0; i < 7; i += 1) {
      const payload = makeRegistrationPayload();
      const response = await registrationApi.startRegistration(payload, {
        "x-device-fingerprint": "test-device-fingerprint-01",
      });
      if (response.status() === 429) {
        blocked = true;
        break;
      }
    }

    expect(blocked).toBe(true);
  });

  test("REG-015: registration flow emits audit and lifecycle events", async ({ registrationApi }) => {
    const payload = makeRegistrationPayload();
    const startResponse = await registrationApi.startRegistration(payload);
    expect([201, 202]).toContain(startResponse.status());
    const startData = (await startResponse.json()) as StartRegistrationResponse;
    const verifyIdentifier = resolveVerifyIdentifier(startData, payload.username);

    const otpCode = await registrationApi.getLatestOtp(verifyIdentifier, payload.phone_number);
    const verifyResponse = await registrationApi.verifyOtp(verifyIdentifier, otpCode);
    expect([200, 201]).toContain(verifyResponse.status());

    if (process.env.PW_TEST_AUDIT_ENDPOINT) {
      const eventsResponse = await registrationApi.getAuditEvents(verifyIdentifier);
      expect(eventsResponse.ok()).toBeTruthy();
      const events = (await eventsResponse.json()) as Array<{ event: string }>;
      const names = events.map((evt) => evt.event);

      expect(names).toContain("registration_started");
      expect(names).toContain("otp_sent");
      expect(names).toContain("otp_verified");
      expect(names).toContain("registration_completed");
      return;
    }

    const statsResponse = await registrationApi.getSMSStats();
    expect(statsResponse.ok()).toBeTruthy();
  });
});
