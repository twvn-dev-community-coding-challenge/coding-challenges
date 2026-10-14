package sms

// Telemetry hooks keep prometheus wiring out of this package (avoids import cycles).

var (
	telemetryMessageCreated func()
	telemetryRecoveryEvent  func(kind string)
)

// SetTelemetryHooks registers optional callbacks for observability. Pass nil to disable a hook.
func SetTelemetryHooks(onMessageCreated func(), onRecoveryEvent func(kind string)) {
	telemetryMessageCreated = onMessageCreated
	telemetryRecoveryEvent = onRecoveryEvent
}

func notifyMessageCreated() {
	if telemetryMessageCreated != nil {
		telemetryMessageCreated()
	}
}

func notifyRecoveryEvent(kind string) {
	if telemetryRecoveryEvent != nil {
		telemetryRecoveryEvent(kind)
	}
}
