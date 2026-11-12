import { SmsProvider } from "./SmsProvider";

/**
 * Registry of available SMS provider adapters.
 * Looked up by name during routing to retrieve the correct adapter instance.
 */
export class ProviderRegistry {
	private providers = new Map<string, SmsProvider>();

	register(provider: SmsProvider): void {
		this.providers.set(provider.name, provider);
	}

	get(name: string): SmsProvider {
		const provider = this.providers.get(name);
		if (!provider) {
			throw new Error(`Provider not registered: ${name}`);
		}
		return provider;
	}

	getAll(): SmsProvider[] {
		return Array.from(this.providers.values());
	}
}
