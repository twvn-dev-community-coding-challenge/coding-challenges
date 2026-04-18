export type ChannelType = 'SMS';

export interface SmsChannelPayload {
  readonly country_code: string;
  readonly phone_number: string;
}

export interface CreateNotificationRequest {
  readonly message_id: string;
  /** Optional; backend defaults to SMS. */
  readonly channel_type?: ChannelType;
  readonly recipient: string;
  readonly content: string;
  readonly channel_payload: SmsChannelPayload;
  /** When true, otp-service generates the code (use {{OTP}} placeholder in content). */
  readonly issue_server_otp?: boolean;
}

export interface DispatchRequest {
  readonly as_of: string;
}

export interface RetryRequest {
  readonly as_of?: string | null;
}

export interface NotificationResource {
  readonly notification_id: string;
  readonly message_id: string;
  readonly channel_type: string;
  readonly recipient: string;
  readonly content: string;
  readonly channel_payload: Record<string, string>;
  readonly state: string;
  readonly attempt: number;
  readonly selected_provider_id: string | null;
  readonly routing_rule_version: number | null;
  readonly estimated_cost?: number | null;
  readonly estimated_currency?: string | null;
  readonly charging_estimate_id?: string | null;
  readonly charging_rate_id?: string | null;
  readonly last_actual_cost?: number | null;
  readonly actual_currency?: string | null;
  readonly charging_actual_cost_id?: string | null;
  readonly otp_challenge_id?: string | null;
  readonly otp_expires_at?: string | null;
  /** Dev-only when server OTP is enabled and OTP_EXPOSE_PLAINTEXT_TO_CLIENT=true */
  readonly otp_plaintext?: string | null;
  readonly created_at: string;
  readonly updated_at: string;
}

export interface ApiErrorBody {
  readonly code: string;
  readonly message: string;
  readonly details?: unknown;
}

export type Result<T, E = ApiErrorBody> =
  | { readonly ok: true; readonly value: T }
  | { readonly ok: false; readonly error: E };

/** Envelope for `GET /notifications`. */
export interface ListNotificationsData {
  readonly notifications: readonly NotificationResource[];
}
