import { test, expect } from "../../fixtures/registration.fixture";
import { makeRegistrationPayload } from "../../helpers/test-data";

test.describe("Registration validation scenarios", () => {
  test("REG-002: missing required fields returns bad request", async ({ registrationApi }) => {
    const payload = makeRegistrationPayload({ username: "" });
    const response = await registrationApi.startRegistration(payload);

    expect(response.status()).toBe(400);
  });

  test("REG-003: duplicate username returns conflict", async ({ registrationApi }) => {
    const duplicateUsername = `dupe_name_${Date.now()}`;
    const firstPayload = makeRegistrationPayload({ username: duplicateUsername });
    const secondPayload = makeRegistrationPayload({ username: duplicateUsername });

    const firstResponse = await registrationApi.startRegistration(firstPayload);
    expect([202, 201]).toContain(firstResponse.status());

    const secondResponse = await registrationApi.startRegistration(secondPayload);
    expect(secondResponse.status()).toBe(409);
  });

  test("REG-004: duplicate phone number returns conflict", async ({ registrationApi }) => {
    const duplicatePhone = `63917${`${Date.now()}`.slice(-7)}`;
    const firstPayload = makeRegistrationPayload({ phone_number: duplicatePhone });
    const secondPayload = makeRegistrationPayload({ phone_number: duplicatePhone });

    const firstResponse = await registrationApi.startRegistration(firstPayload);
    expect([202, 201]).toContain(firstResponse.status());

    const secondResponse = await registrationApi.startRegistration(secondPayload);
    expect(secondResponse.status()).toBe(409);
  });
});
