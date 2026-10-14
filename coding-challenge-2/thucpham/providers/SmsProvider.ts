import { SmsMessage } from "../domain/SmsMessage";

export interface SmsProvider {
	name: string;
	getEstimatedCost(carrier: string): number;
	send(sms: SmsMessage): void;
}
