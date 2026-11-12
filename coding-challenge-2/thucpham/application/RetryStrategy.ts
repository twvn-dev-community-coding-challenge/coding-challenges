import { SmsMessage } from "../domain/SmsMessage";

/**
 * Determines how an SMS should be retried after a delivery failure.
 */
export interface RetryStrategy {
	/** Returns the provider name to use on retry. */
	selectProvider(sms: SmsMessage, availableProviders: string[]): string;

	/**
	 * Providers excluded from routing lookup on retry.
	 * Same-provider retry must not drop the last provider from candidates, or transient retries cannot re-select it.
	 */
	failedProvidersToExcludeFromRouting(sms: SmsMessage): string[];
}

/**
 * Retries using the same provider — suitable for transient errors (e.g. temporary overload).
 */
export class SameProviderStrategy implements RetryStrategy {
	failedProvidersToExcludeFromRouting(sms: SmsMessage): string[] {
		const failed = sms.getFailedProviders();
		const last = sms.getLastUsedProvider();
		if (!last) return failed;
		return failed.filter((p) => p !== last);
	}

	selectProvider(sms: SmsMessage, availableProviders: string[]): string {
		const last = sms.getLastUsedProvider();
		if (last && availableProviders.includes(last)) {
			return last;
		}
		return availableProviders[0];
	}
}

/**
 * Retries using a different provider — suitable for permanent failures (e.g. provider down, carrier block).
 */
export class FallbackProviderStrategy implements RetryStrategy {
	failedProvidersToExcludeFromRouting(sms: SmsMessage): string[] {
		return sms.getFailedProviders();
	}

	selectProvider(sms: SmsMessage, availableProviders: string[]): string {
		const failedProviders = sms.getFailedProviders();
		const candidates = availableProviders.filter((p) => !failedProviders.includes(p));

		if (candidates.length === 0) {
			throw new Error(`No fallback provider available for SMS ${sms.id}`);
		}

		return candidates[0];
	}
}

/**
 * Resolves the appropriate retry strategy based on error code from the provider callback.
 * Transient errors use same-provider retry; permanent errors use fallback provider.
 */
export class RetryStrategyResolver {
	/** Same-provider retry — temporary / throttling. */
	private static readonly TRANSIENT_ERROR_CODES = new Set(["RATE_LIMITED"]);

	static resolve(errorCode?: string): RetryStrategy {
		if (errorCode && RetryStrategyResolver.TRANSIENT_ERROR_CODES.has(errorCode)) {
			return new SameProviderStrategy();
		}
		return new FallbackProviderStrategy();
	}
}
