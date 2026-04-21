export type ClientPhase = 'idle' | 'submitting' | 'sent';

export type SmsScenarioId = 'registration-otp';

export interface MembershipSmsFormValues {
  readonly messageId: string;
  readonly recipient: string;
  readonly countryCode: string;
  readonly phoneNumber: string;
  readonly content: string;
}
