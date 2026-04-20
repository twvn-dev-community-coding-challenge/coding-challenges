import { expect } from "@playwright/test";
import type { StartRegistrationResponse } from "../fixtures/registration.fixture";

export function hasEnv(varName: string): boolean {
  return Boolean(process.env[varName]);
}

export function requireEnv(varName: string): string {
  const value = process.env[varName];
  if (!value) {
    throw new Error(`${varName} must be set for this test`);
  }
  return value;
}

export function resolveVerifyIdentifier(payload: StartRegistrationResponse, fallbackUsername: string): string {
  const verifyIdField = process.env.PW_VERIFY_ID_FIELD || "username";

  if (verifyIdField === "registration_id") {
    expect(payload.registration_id, "registration_id should be returned").toBeTruthy();
    return payload.registration_id as string;
  }

  return fallbackUsername;
}
