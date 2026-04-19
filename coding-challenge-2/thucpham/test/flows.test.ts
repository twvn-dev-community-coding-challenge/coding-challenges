import { describe, it, expect, beforeEach, vi } from "vitest";
import { StateMachine } from "../domain/StateMachine";
import { SmsStatus } from "../domain/SmsStatus";
import { SmsRepository } from "../infrastructure/SmsRepository";
import {
	VonageAdapter,
	TwilioAdapter,
	InfobipAdapter,
	AwsSnsAdapter,
	TelnyxAdapter,
	MessageBirdAdapter,
	SinchAdapter,
} from "../providers/Providers";
import { ProviderRegistry } from "../providers/ProviderRegistry";
import { CarrierDetector } from "../routing/CarrierDetector";
import { RoutingEngine } from "../routing/RoutingEngine";
import { RoutingTable } from "../routing/RoutingTable";
import { CallbackHandler } from "../application/CallbackHandler";
import { CostReport } from "../application/CostReport";
import { SmsService } from "../application/SmsService";

/** Same wiring as `main.ts` — integration-style tests against real objects. */
function buildSystem() {
	const repository = new SmsRepository();
	const stateMachine = new StateMachine();

	const registry = new ProviderRegistry();
	registry.register(new VonageAdapter());
	registry.register(new TwilioAdapter());
	registry.register(new InfobipAdapter());
	registry.register(new AwsSnsAdapter());
	registry.register(new TelnyxAdapter());
	registry.register(new MessageBirdAdapter());
	registry.register(new SinchAdapter());

	const detector = new CarrierDetector();
	const table = new RoutingTable();
	table.addRule("VN", "Viettel", ["Twilio", "Vonage"]);
	table.addRule("VN", "Mobifone", ["Vonage", "Twilio", "Infobip"]);
	table.addRule("VN", "Vinaphone", ["Vonage", "Twilio"]);
	table.addRule("TH", "AIS", ["Infobip", "Twilio"]);
	table.addRule("TH", "DTAC", ["AwsSns", "Infobip"]);
	table.addRule("SG", "Singtel", ["Twilio", "Infobip"]);
	table.addRule("SG", "StarHub", ["Telnyx", "Twilio"]);
	table.addRule("PH", "Globe", ["MessageBird", "Twilio"]);
	table.addRule("PH", "Smart", ["Sinch", "MessageBird"]);
	table.addRule("PH", "DITO", ["MessageBird", "Sinch"]);

	const routing = new RoutingEngine(detector, table);
	const smsService = new SmsService(routing, registry, stateMachine, repository);
	const callbackHandler = new CallbackHandler(repository, stateMachine, smsService);
	const costReport = new CostReport(repository);

	return { repository, smsService, callbackHandler, costReport };
}

describe("SMS flows (integration)", () => {
	let repository: SmsRepository;
	let smsService: SmsService;
	let callbackHandler: CallbackHandler;
	let costReport: CostReport;

	beforeEach(() => {
		vi.spyOn(console, "log").mockImplementation(() => {});
		const s = buildSystem();
		repository = s.repository;
		smsService = s.smsService;
		callbackHandler = s.callbackHandler;
		costReport = s.costReport;
	});

	it("send fails when country has no routing / carrier data", () => {
		expect(() => smsService.send("bad", "XX", "0321234567", "x")).toThrow(/No carrier data/);
	});

	it("send fails when phone cannot be matched to a carrier", () => {
		expect(() => smsService.send("bad2", "VN", "0000000000", "x")).toThrow(/Cannot detect carrier/);
	});

	it("happy path: send then callbacks → Send-success with actual cost", () => {
		const sms = smsService.send("m1", "VN", "0321234567", "OTP 1");
		expect(sms.status).toBe(SmsStatus.SEND_TO_PROVIDER);
		expect(sms.provider).toBe("Twilio");

		callbackHandler.handle("m1", SmsStatus.QUEUE);
		callbackHandler.handle("m1", SmsStatus.SEND_TO_CARRIER);
		callbackHandler.handle("m1", SmsStatus.SEND_SUCCESS, 0.019);

		const saved = repository.findById("m1");
		expect(saved?.status).toBe(SmsStatus.SEND_SUCCESS);
		expect(saved?.actualCost).toBe(0.019);
	});

	it("callback fail + RATE_LIMITED → retry with same provider (Twilio)", () => {
		smsService.send("m2", "VN", "0321234567", "OTP 2");
		callbackHandler.handle("m2", SmsStatus.QUEUE);
		callbackHandler.handle("m2", SmsStatus.SEND_TO_CARRIER);
		callbackHandler.handle("m2", SmsStatus.SEND_FAILED, undefined, "RATE_LIMITED");

		const sms = repository.findById("m2");
		expect(sms?.status).toBe(SmsStatus.SEND_TO_PROVIDER);
		expect(sms?.provider).toBe("Twilio");
	});

	it("callback fail + SERVICE_UNAVAILABLE → retry with fallback provider (Vonage)", () => {
		smsService.send("m3", "VN", "0321234567", "OTP 3");
		callbackHandler.handle("m3", SmsStatus.QUEUE);
		callbackHandler.handle("m3", SmsStatus.SEND_TO_CARRIER);
		callbackHandler.handle("m3", SmsStatus.SEND_FAILED, undefined, "SERVICE_UNAVAILABLE");

		const sms = repository.findById("m3");
		expect(sms?.status).toBe(SmsStatus.SEND_TO_PROVIDER);
		expect(sms?.provider).toBe("Vonage");
	});

	it("repository: list persisted messages", () => {
		smsService.send("a", "VN", "0321000000", "x");
		smsService.send("b", "VN", "0321000001", "y");
		const all = repository.findAll();
		expect(all.length).toBe(2);
		expect(all.map((m) => m.id).sort()).toEqual(["a", "b"]);
	});

	it("cost report: totals by provider and country after a success", () => {
		smsService.send("c1", "VN", "0321111111", "ok");
		callbackHandler.handle("c1", SmsStatus.QUEUE);
		callbackHandler.handle("c1", SmsStatus.SEND_TO_CARRIER);
		callbackHandler.handle("c1", SmsStatus.SEND_SUCCESS, 0.05);

		const twilioCost = costReport.totalCostByProvider("Twilio");
		const vnCost = costReport.totalCostByCountry("VN");
		expect(twilioCost).toBe(0.05);
		expect(vnCost).toBe(0.05);
		expect(costReport.volumeByProvider("Twilio")).toBe(1);
		expect(costReport.successRateByProvider("Twilio")).toBe(1);
	});
});
