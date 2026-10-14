import { SmsMessage } from "../domain/SmsMessage";
import { SmsProvider } from "./SmsProvider";

abstract class BaseAdapter implements SmsProvider {
	abstract name: string;
	abstract readonly costs: Record<string, number>;

	getEstimatedCost(carrier: string): number {
		return this.costs[carrier] ?? 0.01;
	}

	send(sms: SmsMessage): void {
		console.log(`[${this.name}] Sending SMS ${sms.id} to ${sms.phoneNumber} (${sms.carrier})`);
		console.log(`[${this.name}] Estimated cost: $${this.getEstimatedCost(sms.carrier)}`);
	}
}

export class VonageAdapter extends BaseAdapter {
	name = "Vonage";
	readonly costs: Record<string, number> = {
		Viettel: 0.03,
		Mobifone: 0.018,
		Vinaphone: 0.012,
	};
}

export class TwilioAdapter extends BaseAdapter {
	name = "Twilio";
	readonly costs: Record<string, number> = {
		Viettel: 0.015,
		Mobifone: 0.02,
		Singtel: 0.022,
	};
}

export class InfobipAdapter extends BaseAdapter {
	name = "Infobip";
	readonly costs: Record<string, number> = {
		AIS: 0.018,
		Mobifone: 0.017,
	};
}

export class AwsSnsAdapter extends BaseAdapter {
	name = "AwsSns";
	readonly costs: Record<string, number> = {
		DTAC: 0.014,
	};
}

export class TelnyxAdapter extends BaseAdapter {
	name = "Telnyx";
	readonly costs: Record<string, number> = {
		StarHub: 0.019,
	};
}

export class MessageBirdAdapter extends BaseAdapter {
	name = "MessageBird";
	readonly costs: Record<string, number> = {
		Globe: 0.016,
		DITO: 0.016,
	};
}

export class SinchAdapter extends BaseAdapter {
	name = "Sinch";
	readonly costs: Record<string, number> = {
		Smart: 0.015,
	};
}
