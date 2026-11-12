import { SmsMessage } from "../domain/SmsMessage";

/**
 * In-memory store for all SMS messages processed by the system.
 */
export class SmsRepository {
	private store = new Map<string, SmsMessage>();

	save(sms: SmsMessage): void {
		this.store.set(sms.id, sms);
	}

	findById(id: string): SmsMessage | null {
		return this.store.get(id) ?? null;
	}

	findAll(): SmsMessage[] {
		return Array.from(this.store.values());
	}

	findByProvider(provider: string): SmsMessage[] {
		return this.findAll().filter((sms) => sms.provider === provider);
	}

	findByCountry(country: string): SmsMessage[] {
		return this.findAll().filter((sms) => sms.country === country);
	}
}
