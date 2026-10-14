import { randomBytes } from "node:crypto";
import { test as base, APIRequestContext, expect } from "@playwright/test";
import type { RegistrationPayload } from "../helpers/test-data";

type StartRegistrationResponse = {
  message?: string;
  registration_id?: string;
  otp_channel?: string;
};

type VerifyOtpResponse = {
  message?: string;
  user?: {
    id: number;
    username: string;
    phone_number: string;
  };
};

function hasForwardedFor(headers: Record<string, string>): boolean {
  return Object.keys(headers).some((k) => k.toLowerCase() === "x-forwarded-for");
}

export class RegistrationApi {
  constructor(
    private readonly request: APIRequestContext,
    private readonly nextIsolatedClientIP: () => string,
  ) {}

  async startRegistration(payload: Partial<RegistrationPayload>, headers: Record<string, string> = {}) {
    const startEndpoint = process.env.PW_REGISTER_START_ENDPOINT || "/api/register";
    const merged: Record<string, string> = { ...headers };
    if (!hasForwardedFor(merged)) {
      merged["X-Forwarded-For"] = this.nextIsolatedClientIP();
    }
    return this.request.post(startEndpoint, {
      data: payload,
      headers: merged,
    });
  }

  async verifyOtp(registrationId: string, otpCode: string, headers: Record<string, string> = {}) {
    const verifyEndpoint = process.env.PW_REGISTER_VERIFY_ENDPOINT || "/api/verify";
    const verifyIdField = process.env.PW_VERIFY_ID_FIELD || "username";
    return this.request.post(verifyEndpoint, {
      data: {
        [verifyIdField]: registrationId,
        otp_code: otpCode,
      },
      headers,
    });
  }

  async getLatestOtp(registrationId: string, phoneNumber: string): Promise<string> {
    // Fallback for current backend: parse OTP from latest SMS content.
    const messagesResponse = await this.request.get("/api/sms/messages");
    expect(messagesResponse.ok()).toBeTruthy();

    const messages = (await messagesResponse.json()) as Array<{
      phone_number?: string;
      content?: string;
      created_at?: string;
    }>;

    const related = messages
      .filter((msg) => msg.phone_number === phoneNumber && typeof msg.content === "string")
      .sort((a, b) => new Date(b.created_at || 0).getTime() - new Date(a.created_at || 0).getTime());

    const match = related[0]?.content?.match(/(\d{6})/);
    if (!match) {
      throw new Error(`OTP code was not found in SMS messages for phone: ${phoneNumber}`);
    }
    return match[1];
  }

  async getUserByUsername(username: string) {
    const userLookupEndpoint = process.env.PW_TEST_USER_LOOKUP_ENDPOINT;
    if (!userLookupEndpoint) {
      throw new Error("PW_TEST_USER_LOOKUP_ENDPOINT is required for user lookup assertions");
    }
    return this.request.get(userLookupEndpoint, { params: { username } });
  }

  async getAuditEvents(registrationId: string) {
    const auditEndpoint = process.env.PW_TEST_AUDIT_ENDPOINT;
    if (!auditEndpoint) {
      throw new Error("PW_TEST_AUDIT_ENDPOINT is required for audit assertions");
    }
    return this.request.get(auditEndpoint, { params: { registration_id: registrationId } });
  }

  async getSMSStats() {
    return this.request.get("/api/sms/stats");
  }
}

type Fixtures = {
  registrationApi: RegistrationApi;
};

function randomPrivateIPv4(): string {
  const [a, b, c] = randomBytes(3);
  return `10.${a}.${b}.${c}`;
}

export const test = base.extend<Fixtures>({
  registrationApi: async ({ request }, use) => {
    await use(new RegistrationApi(request, randomPrivateIPv4));
  },
});

export { expect };
export type { StartRegistrationResponse, VerifyOtpResponse };
