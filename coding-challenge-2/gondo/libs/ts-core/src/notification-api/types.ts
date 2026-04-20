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
  readonly selected_provider_code?: string | null;
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

/** One row from runtime pipeline aggregation (`GET .../pipeline-events`). */
export interface PipelineEvent {
  readonly timestamp: string;
  readonly phase: string;
  readonly detail: Readonly<Record<string, unknown>>;
}

/** Envelope for `GET /notifications/{id}/pipeline-events`. */
export interface PipelineEventsData {
  readonly notification_id: string;
  readonly message_id: string;
  readonly events: readonly PipelineEvent[];
}

/** One bucket from `GET /notifications/kpis` (in-memory aggregates). */
export interface SmsKpisBucketRow {
  readonly volume: number;
  readonly total_estimated_cost: number;
  readonly total_actual_cost: number;
  readonly with_estimate_count: number;
  readonly with_actual_count: number;
  readonly send_success: number;
  readonly send_failed: number;
  readonly carrier_rejected: number;
  readonly in_flight: number;
  readonly terminal_failure: number;
  readonly terminal_success_rate: number | null;
  readonly terminal_failure_rate: number | null;
}

export interface SmsKpisOverall extends SmsKpisBucketRow {
  readonly total_notifications: number;
}

export interface SmsKpisByProvider extends SmsKpisBucketRow {
  readonly provider_id: string;
  readonly provider_code: string | null;
}

export interface SmsKpisByCountry extends SmsKpisBucketRow {
  readonly country_code: string;
}

export interface SmsKpisByCallingDomain extends SmsKpisBucketRow {
  readonly calling_domain: string;
}

/** Echo of optional `from` / `to` query window on `GET /notifications/kpis`. */
export interface SmsKpisCreatedAtFilter {
  readonly from: string | null;
  readonly to: string | null;
}

/** Response body `data` from `GET /notifications/kpis`. */
export interface SmsKpisData {
  readonly source: string;
  readonly created_at_filter: SmsKpisCreatedAtFilter;
  readonly currency_note: string;
  readonly overall: SmsKpisOverall;
  readonly by_provider: readonly SmsKpisByProvider[];
  readonly by_country: readonly SmsKpisByCountry[];
  readonly by_calling_domain: readonly SmsKpisByCallingDomain[];
}
