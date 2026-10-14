import { test, expect } from "../../fixtures/registration.fixture";
import type { StartRegistrationResponse, VerifyOtpResponse } from "../../fixtures/registration.fixture";
import { makeRegistrationPayload } from "../../helpers/test-data";
import { resolveVerifyIdentifier } from "../../helpers/otp-helper";

test.describe("Registration success scenarios", () => {
  test("REG-001: happy path registration with valid OTP", async ({ registrationApi }) => {
    const payload = makeRegistrationPayload();
    const startResponse = await registrationApi.startRegistration(payload);
    expect([201, 202]).toContain(startResponse.status());

    const startData = (await startResponse.json()) as StartRegistrationResponse;
    const verifyIdentifier = resolveVerifyIdentifier(startData, payload.username);

    const otpCode = await registrationApi.getLatestOtp(verifyIdentifier, payload.phone_number);
    const verifyResponse = await registrationApi.verifyOtp(verifyIdentifier, otpCode);
    expect([200, 201]).toContain(verifyResponse.status());

    const verifyData = (await verifyResponse.json()) as VerifyOtpResponse;
    expect(verifyData.message).toBeTruthy();
  });

  test("REG-014: data integrity after successful verification", async ({ registrationApi }) => {
    const payload = makeRegistrationPayload();
    const startResponse = await registrationApi.startRegistration(payload);
    expect([201, 202]).toContain(startResponse.status());
    const startData = (await startResponse.json()) as StartRegistrationResponse;

    const verifyIdentifier = resolveVerifyIdentifier(startData, payload.username);
    const otpCode = await registrationApi.getLatestOtp(verifyIdentifier, payload.phone_number);
    const verifyResponse = await registrationApi.verifyOtp(verifyIdentifier, otpCode);
    expect([200, 201]).toContain(verifyResponse.status());

    if (process.env.PW_TEST_USER_LOOKUP_ENDPOINT) {
      const userLookupResponse = await registrationApi.getUserByUsername(payload.username);
      expect(userLookupResponse.ok()).toBeTruthy();
      const userData = (await userLookupResponse.json()) as {
        username?: string;
        phone_number?: string;
        password_hash?: string;
        password?: string;
      };

      expect(userData.username).toBe(payload.username);
      expect(userData.phone_number).toBe(payload.phone_number);
      expect(userData.password_hash, "password should be stored as hash").toBeTruthy();
      expect(userData.password, "plain password should never be persisted").toBeFalsy();
      return;
    }

    // Fallback assertion for current backend without test-only lookup endpoint:
    // uniqueness checks prove user persistence after successful verification.
    const dupeUsernameResponse = await registrationApi.startRegistration(
      makeRegistrationPayload({
        username: payload.username,
      }),
    );
    expect(dupeUsernameResponse.status()).toBe(409);
  });
});
