import * as p from "@clack/prompts";
import { SmsMessage } from "./domain/SmsMessage";
import { SmsStatus } from "./domain/SmsStatus";
import { CallbackHandler } from "./application/CallbackHandler";
import { CostReport } from "./application/CostReport";
import { SmsService } from "./application/SmsService";
import { SmsRepository } from "./infrastructure/SmsRepository";

const COUNTRIES = ["VN", "SG", "TH", "PH"] as const;

/** Sample numbers whose prefixes match CarrierDetector for each country. */
const DEFAULT_PHONE_BY_COUNTRY: Record<(typeof COUNTRIES)[number], string> = {
	VN: "0912345678", // Vinaphone (091)
	TH: "0812345678", // AIS (081)
	SG: "81234567", // Singtel (leading 8)
	PH: "09171234567", // Globe (0917)
};

/**
 * Interactive command-line interface for sending SMS and observing the lifecycle.
 */
export class CLI {
	private smsService: SmsService;
	private callbackHandler: CallbackHandler;
	private costReport: CostReport;
	private repository: SmsRepository;

	constructor(
		smsService: SmsService,
		callbackHandler: CallbackHandler,
		costReport: CostReport,
		repository: SmsRepository,
	) {
		this.smsService = smsService;
		this.callbackHandler = callbackHandler;
		this.costReport = costReport;
		this.repository = repository;
	}

	async start(): Promise<void> {
		p.intro("  SMS Gateway CLI  ");

		while (true) {
			const action = await p.select({
				message: "What do you want to do?",
				options: [
					{ value: "send", label: "Send SMS" },
					{ value: "callback", label: "Simulate provider callback" },
					{ value: "list", label: "View all SMS" },
					{ value: "report", label: "Cost report" },
					{ value: "exit", label: "Exit" },
				],
			});

			if (p.isCancel(action) || action === "exit") {
				p.outro("Bye.");
				process.exit(0);
			}

			switch (action) {
				case "send":
					await this.handleSend();
					break;
				case "callback":
					await this.handleCallback();
					break;
				case "list":
					await this.handleListAll();
					break;
				case "report":
					await this.handleReport();
					break;
			}
		}
	}

	/** Human-readable status with emoji for terminal / notable states. */
	private displayStatus(status: SmsStatus): string {
		switch (status) {
			case SmsStatus.SEND_SUCCESS:
				return `✅ ${status}`;
			case SmsStatus.SEND_FAILED:
				return `❌ ${status}`;
			case SmsStatus.CARRIER_REJECTED:
				return `🚫 ${status}`;
			case SmsStatus.NEW:
				return `📋 ${status}`;
			case SmsStatus.SEND_TO_PROVIDER:
				return `📤 ${status}`;
			case SmsStatus.QUEUE:
				return `⏳ ${status}`;
			case SmsStatus.SEND_TO_CARRIER:
				return `📡 ${status}`;
			default:
				return String(status);
		}
	}

	/** Next `msg-{n}` id so repeated sends do not overwrite the same repository key by default. */
	private nextDefaultMessageId(): string {
		let max = 0;
		for (const sms of this.repository.findAll()) {
			const m = /^msg-(\d+)$/i.exec(sms.id.trim());
			if (m) max = Math.max(max, parseInt(m[1], 10));
		}
		return `msg-${max + 1}`;
	}

	private async handleSend(): Promise<void> {
		const suggestedId = this.nextDefaultMessageId();
		const answers = await p.group(
			{
				id: () =>
					p.text({
						message: "Message ID",
						placeholder: suggestedId,
						defaultValue: suggestedId,
					}),
				country: () =>
					p.select({
						message: "Country",
						options: COUNTRIES.map((c) => ({ value: c, label: c })),
					}),
				phone: ({ results }) => {
					const c = results.country as (typeof COUNTRIES)[number] | undefined;
					const def = c ? DEFAULT_PHONE_BY_COUNTRY[c] : DEFAULT_PHONE_BY_COUNTRY.VN;
					return p.text({
						message: "Phone number",
						placeholder: def,
						defaultValue: def,
					});
				},
				message: () =>
					p.text({
						message: "Message content",
						placeholder: "OTP: 1234",
						defaultValue: "OTP: 1234",
					}),
			},
			{
				onCancel: () => {
					p.cancel("Cancelled.");
					return;
				},
			},
		);

		if (p.isCancel(answers)) return;

		const s = p.spinner();
		s.start("Sending SMS...");

		try {
			const sms = this.smsService.send(
				answers.id as string,
				answers.country as string,
				answers.phone as string,
				answers.message as string,
			);
			s.stop(`SMS sent`);
			p.note(this.formatSmsNote(sms), `SMS ${sms.id}`);
		} catch (err) {
			s.stop("Failed");
			p.log.error(err instanceof Error ? err.message : String(err));
		}
	}

	/** Newest activity first: last audit entry time, then id (numeric-aware). */
	private sortSmsNewestFirst(list: SmsMessage[]): SmsMessage[] {
		const lastTs = (m: SmsMessage) =>
			m.auditLog.length > 0 ? m.auditLog[m.auditLog.length - 1].timestamp.getTime() : 0;
		return [...list].sort((a, b) => {
			const diff = lastTs(b) - lastTs(a);
			if (diff !== 0) return diff;
			return b.id.localeCompare(a.id, undefined, { numeric: true });
		});
	}

	private async handleCallback(): Promise<void> {
		const allSms = this.sortSmsNewestFirst(this.repository.findAll());
		if (allSms.length === 0) {
			p.log.warn("No SMS messages found. Send one first.");
			return;
		}

		const colPick = this.smsListColumnWidths(allSms);
		const id = await p.select({
			message: "Select SMS",
			options: allSms.map((sms) => ({
				value: sms.id,
				label: this.smsListRowLabel(sms, colPick),
			})),
		});
		if (p.isCancel(id)) return;

		const outcome = await p.select({
			message: "Delivery outcome",
			options: [
				{
					value: "success",
					label: "✅ Success (simulate happy path: Queue → Send-to-carrier → Send-success)",
				},
				{
					value: "failed",
					label:
						"❌ Failed (simulate failed path: Queue → Send-to-carrier → Send-failed; will retry)",
				},
			],
		});
		if (p.isCancel(outcome)) return;

		let errorCode: string | undefined;
		if (outcome === "failed") {
			const code = await p.select({
				message: "Failure reason (controls retry strategy)",
				options: [
					{
						value: "RATE_LIMITED",
						label: "RATE_LIMITED (temporary / throttling — will retry with the same provider)",
					},
					{
						value: "SERVICE_UNAVAILABLE",
						label:
							"SERVICE_UNAVAILABLE (provider down — will retry with another fallback provider)",
					},
				],
			});
			if (p.isCancel(code)) return;
			errorCode = code as string;
		}

		let actualCost: number | undefined;
		if (outcome === "success") {
			const sms = this.repository.findById(id as string);
			const estimated = sms?.estimatedCost ?? 0.015;
			// actual cost = estimated ± up to 10%
			const delta = estimated * (Math.random() * 0.1 - 0.05);
			actualCost = parseFloat((estimated + delta).toFixed(4));
		}

		try {
			const smsId = id as string;
			if (outcome === "success") {
				this.callbackHandler.handle(smsId, SmsStatus.QUEUE);
				this.callbackHandler.handle(smsId, SmsStatus.SEND_TO_CARRIER);
				this.callbackHandler.handle(smsId, SmsStatus.SEND_SUCCESS, actualCost);
				p.log.success(`✅ Delivered  →  actual cost $${actualCost}`);
			} else {
				this.callbackHandler.handle(smsId, SmsStatus.QUEUE);
				this.callbackHandler.handle(smsId, SmsStatus.SEND_TO_CARRIER);
				this.callbackHandler.handle(smsId, SmsStatus.SEND_FAILED, undefined, errorCode);
				p.log.warn(`❌ Failed (${errorCode}) — retry triggered`);
			}

			const updated = this.repository.findById(smsId);
			if (updated) {
				p.note(this.formatSmsNote(updated), `SMS ${updated.id}`);
			}
		} catch (err) {
			p.log.error(err instanceof Error ? err.message : String(err));
		}
	}

	/** Column widths derived from the current batch so rows align in `p.select`. */
	private smsListColumnWidths(all: SmsMessage[]): {
		id: number;
		status: number;
		country: number;
		carrier: number;
		provider: number;
	} {
		const st = (s: SmsMessage) => this.displayStatus(s.status);
		return {
			id: Math.max(6, ...all.map((s) => s.id.length)),
			status: Math.max(22, ...all.map((s) => st(s).length)),
			country: Math.max(4, ...all.map((s) => s.country.length)),
			carrier: Math.max(8, ...all.map((s) => s.carrier.length)),
			provider: Math.max(9, ...all.map((s) => s.provider.length)),
		};
	}

	/** One aligned row: id, status (with emoji), country, carrier, provider, est, act. */
	private smsListRowLabel(
		sms: SmsMessage,
		w: { id: number; status: number; country: number; carrier: number; provider: number },
	): string {
		const est = `est $${sms.estimatedCost.toFixed(3)}`;
		const act = `act $${sms.actualCost.toFixed(3)}`;
		return [
			sms.id.padEnd(w.id),
			this.displayStatus(sms.status).padEnd(w.status),
			sms.country.padEnd(w.country),
			sms.carrier.padEnd(w.carrier),
			sms.provider.padEnd(w.provider),
			est.padStart(10),
			act.padStart(10),
		].join("  ");
	}

	private async handleListAll(): Promise<void> {
		while (true) {
			const all = this.sortSmsNewestFirst(this.repository.findAll());
			if (all.length === 0) {
				p.log.warn("No SMS messages found.");
				return;
			}

			const col = this.smsListColumnWidths(all);
			const id = await p.select({
				message: "SMS messages — pick one for full detail",
				options: [
					...all.map((sms) => ({
						value: sms.id,
						label: this.smsListRowLabel(sms, col),
					})),
					{ value: "__back", label: "← Back to menu" },
				],
			});

			if (p.isCancel(id) || id === "__back") return;

			const sms = this.repository.findById(id as string);
			if (!sms) {
				p.log.error("SMS not found");
				continue;
			}

			p.note(this.formatSmsNote(sms), `SMS ${sms.id}`);
		}
	}

	private formatSmsNote(sms: SmsMessage): string {
		const auditLines = sms.auditLog.map((r) => {
			const time = r.timestamp.toISOString().substring(11, 23);
			const transition = `${this.displayStatus(r.fromStatus)} → ${this.displayStatus(r.toStatus)}`;
			return { time, transition, provider: r.provider, carrier: r.carrier };
		});
		const maxLen =
			auditLines.length > 0 ? Math.max(...auditLines.map((l) => l.transition.length)) : 0;
		const formattedAudit = auditLines.map(
			({ time, transition, provider, carrier }) =>
				`[${time}]  ${transition.padEnd(maxLen)}  ${provider}  ·  ${carrier}`,
		);

		return [
			`ID          : ${sms.id}`,
			`Country     : ${sms.country}`,
			`Phone       : ${sms.phoneNumber}`,
			`Message     : ${sms.message}`,
			``,
			`Status      : ${this.displayStatus(sms.status)}`,
			`Provider    : ${sms.provider}`,
			`Carrier     : ${sms.carrier}`,
			`Est. cost   : $${sms.estimatedCost}`,
			`Act. cost   : $${sms.actualCost}`,
			``,
			`Audit log (${sms.auditLog.length} entries):`,
			...formattedAudit,
		].join("\n");
	}

	private async handleReport(): Promise<void> {
		const providers = ["Vonage", "Twilio", "Infobip", "AwsSns", "Telnyx", "MessageBird", "Sinch"];

		// Provider table
		const COL = { name: 12, cost: 10, vol: 8, rate: 10 };
		const hr = `${"─".repeat(COL.name)}┼${"─".repeat(COL.cost)}┼${"─".repeat(COL.vol)}┼${"─".repeat(COL.rate)}`;
		const header = `${"Provider".padEnd(COL.name)}│${"Cost ($)".padStart(COL.cost)}│${"Volume".padStart(COL.vol)}│${"Success".padStart(COL.rate)}`;
		const providerRows = providers.map((name) => {
			const cost = this.costReport.totalCostByProvider(name);
			const volume = this.costReport.volumeByProvider(name);
			const success = (this.costReport.successRateByProvider(name) * 100).toFixed(1) + "%";
			return `${name.padEnd(COL.name)}│${("$" + cost.toFixed(4)).padStart(COL.cost)}│${String(volume).padStart(COL.vol)}│${success.padStart(COL.rate)}`;
		});

		// Country table
		const COL2 = { name: 10, cost: 10 };
		const hr2 = `${"─".repeat(COL2.name)}┼${"─".repeat(COL2.cost)}`;
		const header2 = `${"Country".padEnd(COL2.name)}│${"Cost ($)".padStart(COL2.cost)}`;
		const countryRows = COUNTRIES.map((c) => {
			const cost = this.costReport.totalCostByCountry(c);
			return `${c.padEnd(COL2.name)}│${("$" + cost.toFixed(4)).padStart(COL2.cost)}`;
		});

		p.note(
			[
				"By Provider:",
				"",
				header,
				hr,
				...providerRows,
				"",
				"",
				"By Country:",
				"",
				header2,
				hr2,
				...countryRows,
			].join("\n"),
			"Cost Report",
		);
	}
}
