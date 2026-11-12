import { StateMachine } from "./domain/StateMachine";
import { SmsRepository } from "./infrastructure/SmsRepository";
import {
	VonageAdapter,
	TwilioAdapter,
	InfobipAdapter,
	AwsSnsAdapter,
	TelnyxAdapter,
	MessageBirdAdapter,
	SinchAdapter,
} from "./providers/Providers";
import { ProviderRegistry } from "./providers/ProviderRegistry";
import { CarrierDetector } from "./routing/CarrierDetector";
import { RoutingEngine } from "./routing/RoutingEngine";
import { RoutingTable } from "./routing/RoutingTable";
import { CallbackHandler } from "./application/CallbackHandler";
import { CostReport } from "./application/CostReport";
import { SmsService } from "./application/SmsService";
import { CLI } from "./cli";

// --- Infrastructure ---
const repository = new SmsRepository();
const stateMachine = new StateMachine();

// --- Providers ---
const registry = new ProviderRegistry();
registry.register(new VonageAdapter());
registry.register(new TwilioAdapter());
registry.register(new InfobipAdapter());
registry.register(new AwsSnsAdapter());
registry.register(new TelnyxAdapter());
registry.register(new MessageBirdAdapter());
registry.register(new SinchAdapter());

// --- Routing ---
const detector = new CarrierDetector();
const table = new RoutingTable();

// VN
table.addRule("VN", "Viettel", ["Twilio", "Vonage"]);
table.addRule("VN", "Mobifone", ["Vonage", "Twilio", "Infobip"]);
table.addRule("VN", "Vinaphone", ["Vonage", "Twilio"]);

// TH
table.addRule("TH", "AIS", ["Infobip", "Twilio"]);
table.addRule("TH", "DTAC", ["AwsSns", "Infobip"]);

// SG
table.addRule("SG", "Singtel", ["Twilio", "Infobip"]);
table.addRule("SG", "StarHub", ["Telnyx", "Twilio"]);

// PH
table.addRule("PH", "Globe", ["MessageBird", "Twilio"]);
table.addRule("PH", "Smart", ["Sinch", "MessageBird"]);
table.addRule("PH", "DITO", ["MessageBird", "Sinch"]);

const routing = new RoutingEngine(detector, table);

// --- Application ---
const smsService = new SmsService(routing, registry, stateMachine, repository);
const callbackHandler = new CallbackHandler(repository, stateMachine, smsService);
const costReport = new CostReport(repository);

// --- CLI ---
const cli = new CLI(smsService, callbackHandler, costReport, repository);
cli.start().catch(console.error);
