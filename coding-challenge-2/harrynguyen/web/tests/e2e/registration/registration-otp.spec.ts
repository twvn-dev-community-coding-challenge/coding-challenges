import { test, expect } from "../../fixtures/registration.fixture";
import type { StartRegistrationResponse } from "../../fixtures/registration.fixture";
import { makeRegistrationPayload } from "../../helpers/test-data";
import { resolveVerifyIdentifier } from "../../helpers/otp-helper";

test.describe("Registration OTP and provider behavior", () => {
  test("REG-005: transient provider failure then success via retry or fallback", async ({ registrationApi }) => {
    const payload = makeRegistrationPayload();
    const startResponse = await registrationApi.startRegistration(payload, {
      "x-test-sms-mode": "transient-failure-once",
    });
    expect([201, 202]).toContain(startResponse.status());

    const startData = (await startResponse.json()) as StartRegistrationResponse;
    const verifyIdentifier = resolveVerifyIdentifier(startData, payload.username);
    const otpCode = await registrationApi.getLatestOtp(verifyIdentifier, payload.phone_number);

    const verifyResponse = await registrationApi.verifyOtp(verifyIdentifier, otpCode);
    expect([200, 201]).toContain(verifyResponse.status());
  });

  test("REG-006: hard provider failure returns server error", async ({ registrationApi }) => {
    const payload = makeRegistrationPayload();

    const response = await registrationApi.startRegistration(payload, {
      "x-test-sms-mode": "always-fail",
    });
    expect(response.status()).toBe(500);
  });

  test("REG-007: incorrect OTP returns unauthorized", async ({ registrationApi }) => {
    const payload = makeRegistrationPayload();

    const startResponse = await registrationApi.startRegistration(payload);
    expect([201, 202]).toContain(startResponse.status());
    const startData = (await startResponse.json()) as StartRegistrationResponse;
    const verifyIdentifier = resolveVerifyIdentifier(startData, payload.username);

    const verifyResponse = await registrationApi.verifyOtp(verifyIdentifier, "000000");
    expect(verifyResponse.status()).toBe(401);
  });

  test("REG-008: expired OTP returns gone", async ({ registrationApi }) => {
    const payload = makeRegistrationPayload({
      otp_ttl_seconds: 1,
    });

    const startResponse = await registrationApi.startRegistration(payload);
    expect([201, 202]).toContain(startResponse.status());
    const startData = (await startResponse.json()) as StartRegistrationResponse;
    const verifyIdentifier = resolveVerifyIdentifier(startData, payload.username);

    const expiredCode = await registrationApi.getLatestOtp(verifyIdentifier, payload.phone_number);
    await new Promise((resolve) => setTimeout(resolve, 1200));
    const verifyResponse = await registrationApi.verifyOtp(verifyIdentifier, expiredCode);
    if (verifyResponse.status() !== 410) {
      const body = await verifyResponse.text();
      throw new Error(`Expected 410 for expired OTP, got ${verifyResponse.status()}. Ensure backend is restarted with OTP expiry changes. Response: ${body}`);
    }
  });

  test("REG-009: OTP cannot be reused after successful verification", async ({ registrationApi }) => {
    const payload = makeRegistrationPayload();

    const startResponse = await registrationApi.startRegistration(payload);
    expect([201, 202]).toContain(startResponse.status());
    const startData = (await startResponse.json()) as StartRegistrationResponse;
    const verifyIdentifier = resolveVerifyIdentifier(startData, payload.username);
    const otpCode = await registrationApi.getLatestOtp(verifyIdentifier, payload.phone_number);

    const firstVerify = await registrationApi.verifyOtp(verifyIdentifier, otpCode);
    expect([200, 201]).toContain(firstVerify.status());

    const secondVerify = await registrationApi.verifyOtp(verifyIdentifier, otpCode);
    expect([400, 401, 409]).toContain(secondVerify.status());
  });

  test("REG-013: consumed registration session cannot be verified again", async ({ registrationApi }) => {
    const payload = makeRegistrationPayload();

    const startResponse = await registrationApi.startRegistration(payload);
    expect([201, 202]).toContain(startResponse.status());
    const startData = (await startResponse.json()) as StartRegistrationResponse;
    const verifyIdentifier = resolveVerifyIdentifier(startData, payload.username);
    const otpCode = await registrationApi.getLatestOtp(verifyIdentifier, payload.phone_number);

    const successVerify = await registrationApi.verifyOtp(verifyIdentifier, otpCode);
    expect([200, 201]).toContain(successVerify.status());

    const consumedVerify = await registrationApi.verifyOtp(verifyIdentifier, "123456");
    expect([400, 401, 409]).toContain(consumedVerify.status());
  });
});
