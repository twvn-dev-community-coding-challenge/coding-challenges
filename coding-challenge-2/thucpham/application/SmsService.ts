import { SmsMessage } from "../domain/SmsMessage";
import { SmsStatus } from "../domain/SmsStatus";
import { StateMachine } from "../domain/StateMachine";
import { SmsRepository } from "../infrastructure/SmsRepository";
import { ProviderRegistry } from "../providers/ProviderRegistry";
import { RoutingEngine } from "../routing/RoutingEngine";
import { RetryStrategy } from "./RetryStrategy";

/**
 * Orchestrates the SMS sending flow: routing, provider selection, lifecycle transitions, and persistence.
 */
export class SmsService {
	private routing: RoutingEngine;
	private registry: ProviderRegistry;
	private stateMachine: StateMachine;
	private repository: SmsRepository;

	constructor(
		routing: RoutingEngine,
		registry: ProviderRegistry,
		stateMachine: StateMachine,
		repository: SmsRepository,
	) {
		this.routing = routing;
		this.registry = registry;
		this.stateMachine = stateMachine;
		this.repository = repository;
	}

	send(id: string, country: string, phoneNumber: string, message: string): SmsMessage {
		const sms = new SmsMessage(id, country, phoneNumber, message);

		const carrier = this.routing.resolveCarrier(phoneNumber, country);
		const selectedProvider = this.routing.resolveProvider(phoneNumber, country);
		const adapter = this.registry.get(selectedProvider);

		sms.carrier = carrier;
		sms.provider = selectedProvider;
		sms.estimatedCost = adapter.getEstimatedCost(carrier);
		this.stateMachine.transition(sms, SmsStatus.SEND_TO_PROVIDER);

		adapter.send(sms);
		this.repository.save(sms);
		return sms;
	}

	retry(sms: SmsMessage, strategy: RetryStrategy): void {
		const excludeFromRouting = strategy.failedProvidersToExcludeFromRouting(sms);
		const candidates = this.routing.resolveProviders(
			sms.phoneNumber,
			sms.country,
			excludeFromRouting,
		);
		const selectedProvider = strategy.selectProvider(sms, candidates);
		const adapter = this.registry.get(selectedProvider);

		sms.provider = selectedProvider;
		sms.estimatedCost = adapter.getEstimatedCost(sms.carrier);
		this.stateMachine.transition(sms, SmsStatus.SEND_TO_PROVIDER);

		adapter.send(sms);
		this.repository.save(sms);
	}
}
