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
