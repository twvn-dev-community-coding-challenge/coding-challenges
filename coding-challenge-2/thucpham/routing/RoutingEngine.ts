import { CarrierDetector } from "./CarrierDetector";
import { RoutingTable } from "./RoutingTable";

/**
 * Selects the appropriate SMS provider for a given phone number and country
 * by combining carrier detection with routing table lookup.
 */
export class RoutingEngine {
	private detector: CarrierDetector;
	private table: RoutingTable;

	constructor(detector: CarrierDetector, table: RoutingTable) {
		this.detector = detector;
		this.table = table;
	}

	resolveCarrier(phoneNumber: string, country: string): string {
		return this.detector.detect(phoneNumber, country);
	}

	resolveProvider(phoneNumber: string, country: string): string {
		return this.resolveProviders(phoneNumber, country)[0];
	}

	resolveProviders(
		phoneNumber: string,
		country: string,
		excludeProviders: string[] = [],
	): string[] {
		const carrier = this.resolveCarrier(phoneNumber, country);
		const providers = this.table.lookup(country, carrier);
		const available = providers.filter((p) => !excludeProviders.includes(p));

		if (available.length === 0) {
			throw new Error(
				`No available provider for ${country} + ${carrier} (excluded: ${excludeProviders.join(", ")})`,
			);
		}

		return available;
	}
}
