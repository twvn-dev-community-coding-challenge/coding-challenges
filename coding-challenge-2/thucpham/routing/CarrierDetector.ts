/**
 * Detects the mobile carrier from a phone number using prefix matching.
 */
export class CarrierDetector {
	private static readonly PREFIXES: Record<string, Record<string, string>> = {
		VN: {
			"032": "Viettel",
			"033": "Viettel",
			"034": "Viettel",
			"035": "Viettel",
			"036": "Viettel",
			"037": "Viettel",
			"038": "Viettel",
			"039": "Viettel",
			"086": "Viettel",
			"096": "Viettel",
			"097": "Viettel",
			"098": "Viettel",
			"070": "Mobifone",
			"076": "Mobifone",
			"077": "Mobifone",
			"078": "Mobifone",
			"079": "Mobifone",
			"089": "Mobifone",
			"090": "Mobifone",
			"093": "Mobifone",
			"081": "Vinaphone",
			"082": "Vinaphone",
			"083": "Vinaphone",
			"084": "Vinaphone",
			"085": "Vinaphone",
			"088": "Vinaphone",
			"091": "Vinaphone",
			"094": "Vinaphone",
		},
		TH: {
			"066": "AIS",
			"081": "AIS",
			"082": "AIS",
			"083": "AIS",
			"084": "AIS",
			"085": "AIS",
			"086": "AIS",
			"087": "AIS",
			"088": "AIS",
			"089": "AIS",
			"061": "DTAC",
			"062": "DTAC",
			"063": "DTAC",
			"064": "DTAC",
			"065": "DTAC",
		},
		SG: {
			"8": "Singtel",
			"9": "StarHub",
		},
		PH: {
			"0917": "Globe",
			"0918": "Globe",
			"0926": "Globe",
			"0927": "Globe",
			"0905": "Smart",
			"0907": "Smart",
			"0908": "Smart",
			"0910": "Smart",
			"0991": "DITO",
			"0992": "DITO",
			"0993": "DITO",
		},
	};

	detect(phoneNumber: string, country: string): string {
		const prefixMap = CarrierDetector.PREFIXES[country];
		if (!prefixMap) {
			throw new Error(`No carrier data for country: ${country}`);
		}

		const digits = phoneNumber.replace(/\D/g, "");

		for (const prefix of Object.keys(prefixMap).sort((a, b) => b.length - a.length)) {
			if (digits.startsWith(prefix)) {
				return prefixMap[prefix];
			}
		}

		throw new Error(`Cannot detect carrier for number: ${phoneNumber} in ${country}`);
	}
}
