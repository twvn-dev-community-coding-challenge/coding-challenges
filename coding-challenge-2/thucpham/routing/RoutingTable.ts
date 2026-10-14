/**
 * Stores the mapping of (country, carrier) → ordered list of providers.
 * First entry is the primary provider; subsequent entries are fallbacks.
 * Rules can be added or removed at runtime without touching any other class.
 */
export class RoutingTable {
	private rules = new Map<string, string[]>();

	private key(country: string, carrier: string): string {
		return `${country}|${carrier}`;
	}

	addRule(country: string, carrier: string, providers: string | string[]): void {
		const list = Array.isArray(providers) ? providers : [providers];
		this.rules.set(this.key(country, carrier), list);
	}

	removeRule(country: string, carrier: string): void {
		this.rules.delete(this.key(country, carrier));
	}

	lookup(country: string, carrier: string): string[] {
		return this.rules.get(this.key(country, carrier)) ?? [];
	}
}
