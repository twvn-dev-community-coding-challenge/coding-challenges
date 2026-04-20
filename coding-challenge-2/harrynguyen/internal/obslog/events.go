// Central definitions for structured log messages and JSON fields (Loki / Grafana filters).
package obslog

// Slog message (JSON field "msg") — stable identifiers for filtering.
const (
	MsgHTTPRequest               = "http_request"
	MsgSMSRecoveryEvent          = "sms_recovery_event"
	MsgRegisterFailed            = "register_failed"
	MsgRegisterRejected          = "register_rejected"
	MsgRegisterRollbackFailed    = "register_rollback_failed"
	MsgRegisterRolledBack        = "register_rolled_back"
	MsgRegisterSMSSent           = "register_sms_sent"
	MsgRegistrationOTP           = "registration_otp"
	MsgVerifyOTPFailed           = "verify_otp_failed"
	MsgVerifyOTPRejected         = "verify_otp_rejected"
	MsgSMSCallbackBindFailed     = "sms_callback_bind_failed"
	MsgSMSCallbackValidateFailed = "sms_callback_validate_failed"
	MsgSMSCallbackLookupFailed   = "sms_callback_lookup_failed"
	MsgSMSCallbackDispatch       = "sms_callback_dispatch"
	MsgSMSCallbackApplyFailed    = "sms_callback_apply_failed"
	MsgSMSCallbackApplied        = "sms_callback_applied"
	MsgStatusLogWriteFailed      = "status_log_write_failed"
)

// JSON field "event" (domain grouping).
const (
	EventAuthFlow    = "auth_flow"
	EventSMSCallback = "sms_callback"
	EventSMSRecovery = "sms_recovery"
)

// JSON field "flow" (auth / callback pipeline).
const (
	FlowRegister    = "register"
	FlowVerifyOTP   = "verify_otp"
	FlowSMSCallback = "sms_callback"
)

// JSON field "step" for auth_flow and sms_callback logs.
const (
	StepParseBody          = "parse_body"
	StepValidateRequired   = "validate_required"
	StepNormalizeCountry   = "normalize_country"
	StepValidatePhone      = "validate_phone"
	StepRateLimitIP        = "rate_limit_ip"
	StepRateLimitDevice    = "rate_limit_device"
	StepDBLookupUsername   = "db_lookup_username"
	StepConflictUsername   = "conflict_username"
	StepDBLookupPhone      = "db_lookup_phone"
	StepConflictPhone      = "conflict_phone"
	StepHashPassword       = "hash_password"
	StepDBCreateUser       = "db_create_user"
	StepSMSSendOTP         = "sms_send_otp"
	StepRollbackDeleteUser = "rollback_delete_user"
	StepSMSSent            = "sms_sent"
	StepUserNotFound       = "user_not_found"
	StepAlreadyActive      = "already_active"
	StepOTPExpired         = "otp_expired"
	StepOTPMismatch        = "otp_mismatch"
	StepDBUpdateUser       = "db_update_user"
	StepBind               = "bind"
	StepValidate           = "validate"
	StepCallbackLookup     = "lookup"
	StepCallbackDispatch   = "dispatch"
	StepCallbackApply      = "apply"
	StepCallbackApplied    = "applied"
)

// Rollback / business reasons on structured logs.
const (
	ReasonSMSSendFailed = "sms_send_failed"
)

// RecoveryKind is logged as recovery_kind and used for Prometheus labels (SMS recovery paths).
const (
	RecoveryKindMitigation                      = "mitigation"
	RecoveryKindMappedCarrierReject             = "mapped_carrier_reject"
	RecoveryKindRecoveredQueueToDelivered       = "recovered_queue_to_delivered"
	RecoveryKindRecoveryFailedMissingProviderID = "recovery_failed_missing_provider_id"
	RecoveryKindRecoveredProviderToDelivered    = "recovered_provider_to_delivered"
	RecoveryKindUnrecoverable                   = "unrecoverable"
)
