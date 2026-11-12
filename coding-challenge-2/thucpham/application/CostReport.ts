import { SmsStatus } from "../domain/SmsStatus";
import { SmsRepository } from "../infrastructure/SmsRepository";

/**
 * Provides cost and volume analytics over processed SMS messages.
 */
export class CostReport {
	private repository: SmsRepository;

	constructor(repository: SmsRepository) {
		this.repository = repository;
	}

	totalCostByProvider(provider: string): number {
		return this.repository
			.findByProvider(provider)
			.filter((sms) => sms.status === SmsStatus.SEND_SUCCESS)
			.reduce((sum, sms) => sum + sms.actualCost, 0);
	}

	totalCostByCountry(country: string): number {
		return this.repository
			.findByCountry(country)
			.filter((sms) => sms.status === SmsStatus.SEND_SUCCESS)
			.reduce((sum, sms) => sum + sms.actualCost, 0);
	}

	volumeByProvider(provider: string): number {
		return this.repository.findByProvider(provider).length;
	}

	successRateByProvider(provider: string): number {
		const all = this.repository.findByProvider(provider);
		if (all.length === 0) return 0;
		const succeeded = all.filter((sms) => sms.status === SmsStatus.SEND_SUCCESS).length;
		return succeeded / all.length;
	}
}
