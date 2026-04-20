import { randomInt } from "node:crypto";

export type RegistrationPayload = {
  username: string;
  password: string;
  phone_number: string;
  country: string;
  otp_ttl_seconds?: number;
};

function uniqueSuffix(): string {
  return `${Date.now()}_${randomInt(10_000, 99_999)}_${randomInt(10_000, 99_999)}`;
}

export function makeRegistrationPayload(overrides: Partial<RegistrationPayload> = {}): RegistrationPayload {
  const suffix = uniqueSuffix();
  const subscriber7 = String(randomInt(1_000_000, 9_999_999));
  return {
    username: `user_${suffix}`,
    password: "S3cret!Pass123",
    phone_number: `63917${subscriber7}`,
    country: "PH",
    ...overrides,
  };
}
