import { SmsAuditRecord } from "./SmsAuditRecord";
import { SmsStatus } from "./SmsStatus";

/**
 * Core domain entity representing a single SMS message and its full lifecycle.
 * Owns its own state and audit log.
 */
export class SmsMessage {
	id: string;
	country: string;
	phoneNumber: string;
	message: string;
	carrier: string;
	provider: string;
	status: SmsStatus;
	estimatedCost: number;
	actualCost: number;
	auditLog: SmsAuditRecord[];

	constructor(id: string, country: string, phoneNumber: string, message: string) {
		this.id = id;
		this.country = country;
		this.phoneNumber = phoneNumber;
		this.message = message;
		this.carrier = "";
		this.provider = "";
		this.status = SmsStatus.NEW;
		this.estimatedCost = 0;
		this.actualCost = 0;
		this.auditLog = [];
	}

	transitionTo(newStatus: SmsStatus): void {
		const fromStatus = this.status;
		this.status = newStatus;
		this.auditLog.push({
			fromStatus,
			toStatus: newStatus,
			provider: this.provider,
			carrier: this.carrier,
			timestamp: new Date(),
		});
	}

	getFailedProviders(): string[] {
		return this.auditLog
			.filter(
				(r) => r.toStatus === SmsStatus.SEND_FAILED || r.toStatus === SmsStatus.CARRIER_REJECTED,
			)
			.map((r) => r.provider);
	}

	getLastUsedProvider(): string | null {
		const records = this.auditLog.filter((r) => r.toStatus === SmsStatus.SEND_TO_PROVIDER);
		return records.length > 0 ? records[records.length - 1].provider : null;
	}
}
