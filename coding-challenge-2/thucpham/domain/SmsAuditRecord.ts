import { SmsStatus } from "./SmsStatus";

/**
 * An immutable record of a single state change, captured for audit and traceability.
 */
export interface SmsAuditRecord {
	fromStatus: SmsStatus;
	toStatus: SmsStatus;
	provider: string;
	carrier: string;
	timestamp: Date;
}
