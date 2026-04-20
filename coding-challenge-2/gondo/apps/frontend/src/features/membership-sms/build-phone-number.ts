/** ISO 3166-1 alpha-2 → international dialing prefix. */
const DIALING_PREFIX_BY_ISO_COUNTRY: Readonly<Record<string, string>> = {
  VN: '+84',
  TH: '+66',
  SG: '+65',
  PH: '+63',
} as const;

const SUPPORTED_MEMBERSHIP_SMS_COUNTRIES = new Set(
  Object.keys(DIALING_PREFIX_BY_ISO_COUNTRY),
);

/** Redacted / masking characters — cannot validate digit rules. */
const PHONE_MASK_PATTERN = /[*•·…]/u;

const INVALID_FOR_COUNTRY_MESSAGE =
  'Enter a valid phone number for the selected country.' as const;

const isValidNationalDigits = (
  isoCountryCode: string,
  digits: string,
): boolean => {
  switch (isoCountryCode) {
    case 'VN':
    case 'TH':
      return (
        (digits.length === 10 && digits.startsWith('0')) ||
        (digits.length === 9 && digits[0] !== '0')
      );
    case 'SG':
      return digits.length === 8 && digits[0] !== '0';
    case 'PH':
      return (
        (digits.length === 11 && digits.startsWith('0')) ||
        (digits.length === 10 && digits[0] !== '0')
      );
    default:
      return false;
  }
};

const isValidE164Digits = (
  isoCountryCode: string,
  allDigits: string,
): boolean => {
  switch (isoCountryCode) {
    case 'VN':
      return allDigits.length === 11 && allDigits.startsWith('84');
    case 'TH':
      return allDigits.length === 11 && allDigits.startsWith('66');
    case 'SG':
      return allDigits.length === 10 && allDigits.startsWith('65');
    case 'PH':
      return allDigits.length === 12 && allDigits.startsWith('63');
    default:
      return false;
  }
};

/**
 * Returns `null` when the value matches national or E.164 rules for the selected country;
 * otherwise returns a user-facing validation message.
 */
export const validatePhoneNumber = (
  isoCountryCode: string,
  rawPhone: string,
): string | null => {
  const trimmed = rawPhone.trim();
  if (!trimmed) {
    return 'Phone number is required';
  }
  if (PHONE_MASK_PATTERN.test(trimmed)) {
    return 'Masked phone numbers cannot be validated';
  }
  if (!SUPPORTED_MEMBERSHIP_SMS_COUNTRIES.has(isoCountryCode)) {
    return 'Unsupported country code';
  }

  if (trimmed.startsWith('+')) {
    const digits = trimmed.replace(/\D/g, '');
    return isValidE164Digits(isoCountryCode, digits)
      ? null
      : INVALID_FOR_COUNTRY_MESSAGE;
  }

  const digits = trimmed.replace(/\D/g, '');
  if (digits.length === 0) {
    return INVALID_FOR_COUNTRY_MESSAGE;
  }
  return isValidNationalDigits(isoCountryCode, digits)
    ? null
    : INVALID_FOR_COUNTRY_MESSAGE;
};

/** Target SEA markets for membership SMS. */
export const MEMBERSHIP_SMS_COUNTRY_OPTIONS: readonly {
  readonly value: string;
  readonly label: string;
}[] = [
  { value: 'VN', label: 'Vietnam (+84)' },
  { value: 'TH', label: 'Thailand (+66)' },
  { value: 'SG', label: 'Singapore (+65)' },
  { value: 'PH', label: 'Philippines (+63)' },
] as const;

const E164_PLACEHOLDER_BY_ISO_COUNTRY: Readonly<Record<string, string>> = {
  VN: '+84901234567',
  TH: '+66812345678',
  SG: '+6591234567',
  PH: '+639171234567',
} as const;

export const getE164PlaceholderForCountry = (isoCountryCode: string): string =>
  E164_PLACEHOLDER_BY_ISO_COUNTRY[isoCountryCode] ?? '+84901234567';

/**
 * Builds E.164-style `phone_number` for the notification API: must contain real digits so
 * the backend can resolve carrier prefixes. Masked strings (e.g. ``+84*****9999``) return
 * ``''``. International input may use spaces; ``+`` forms are normalized to ``+<digits>``.
 */
export const buildSmsPhoneNumber = (isoCountryCode: string, rawPhone: string): string => {
  const trimmed = rawPhone.trim();
  if (!trimmed) {
    return '';
  }
  /** Redacted display placeholders — carrier lookup needs full MSISDN digits. */
  if (/[*•·…]/u.test(trimmed)) {
    return '';
  }

  const prefix = DIALING_PREFIX_BY_ISO_COUNTRY[isoCountryCode];
  if (prefix === undefined) {
    return trimmed;
  }

  let digits = trimmed.replace(/\D/g, '');
  if (digits.length === 0) {
    return '';
  }

  if (trimmed.startsWith('+')) {
    // Already normalized E.164 (+ then digits only)
    if (/^\+[0-9]+$/u.test(trimmed)) {
      return trimmed;
    }
    return `+${digits}`;
  }

  if (digits.startsWith('0')) {
    digits = digits.slice(1);
  }
  return `${prefix}${digits}`;
};
