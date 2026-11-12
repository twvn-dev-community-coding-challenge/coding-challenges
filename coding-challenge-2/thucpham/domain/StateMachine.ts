import { SmsMessage } from "./SmsMessage";
import { SmsStatus } from "./SmsStatus";

/**
 * Enforces valid state transitions for the SMS delivery lifecycle.
 * Single source of truth for which status transitions are allowed.
 */
export class StateMachine {
	private static readonly VALID_TRANSITIONS: [SmsStatus, SmsStatus][] = [
		[SmsStatus.NEW, SmsStatus.SEND_TO_PROVIDER],
		[SmsStatus.SEND_TO_PROVIDER, SmsStatus.QUEUE],
		[SmsStatus.QUEUE, SmsStatus.SEND_TO_CARRIER],
		[SmsStatus.QUEUE, SmsStatus.CARRIER_REJECTED],
		[SmsStatus.SEND_TO_CARRIER, SmsStatus.SEND_SUCCESS],
		[SmsStatus.SEND_TO_CARRIER, SmsStatus.SEND_FAILED],
		// Retry paths
		[SmsStatus.SEND_FAILED, SmsStatus.SEND_TO_PROVIDER],
		[SmsStatus.CARRIER_REJECTED, SmsStatus.SEND_TO_PROVIDER],
	];

	canTransition(from: SmsStatus, to: SmsStatus): boolean {
		return StateMachine.VALID_TRANSITIONS.some(([f, t]) => f === from && t === to);
	}

	transition(sms: SmsMessage, to: SmsStatus): void {
		if (!this.canTransition(sms.status, to)) {
			throw new Error(`Invalid transition: ${sms.status} → ${to}`);
		}
		sms.transitionTo(to);
	}
}
