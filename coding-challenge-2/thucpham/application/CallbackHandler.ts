import { SmsStatus } from "../domain/SmsStatus";
import { StateMachine } from "../domain/StateMachine";
import { SmsRepository } from "../infrastructure/SmsRepository";
import { RetryStrategyResolver } from "./RetryStrategy";
import { SmsService } from "./SmsService";

/**
 * Processes asynchronous callbacks from SMS providers and updates message state accordingly.
 */
export class CallbackHandler {
	private repository: SmsRepository;
	private stateMachine: StateMachine;
	private smsService: SmsService;

	constructor(repository: SmsRepository, stateMachine: StateMachine, smsService: SmsService) {
		this.repository = repository;
		this.stateMachine = stateMachine;
		this.smsService = smsService;
	}

	handle(messageId: string, newState: SmsStatus, actualCost?: number, errorCode?: string): void {
		const sms = this.repository.findById(messageId);
		if (!sms) {
			throw new Error(`SMS not found: ${messageId}`);
		}

		if (newState === SmsStatus.SEND_SUCCESS && actualCost !== undefined) {
			sms.actualCost = actualCost;
		}

		this.stateMachine.transition(sms, newState);

		if (newState === SmsStatus.CARRIER_REJECTED || newState === SmsStatus.SEND_FAILED) {
			const strategy = RetryStrategyResolver.resolve(errorCode);
			this.smsService.retry(sms, strategy);
		}

		this.repository.save(sms);
		console.log(`[Callback] SMS ${messageId} → ${newState}${errorCode ? ` (${errorCode})` : ""}`);
	}
}
