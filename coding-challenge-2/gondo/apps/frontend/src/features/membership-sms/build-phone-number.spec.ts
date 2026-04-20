import { describe, it, expect } from 'vitest';

import {
  buildSmsPhoneNumber,
  getE164PlaceholderForCountry,
  validatePhoneNumber,
} from './build-phone-number';

describe('buildSmsPhoneNumber', () => {
  it('returns empty string when there are no usable digits (caller must validate)', () => {
    expect(buildSmsPhoneNumber('VN', '')).toBe('');
    expect(buildSmsPhoneNumber('VN', '   ')).toBe('');
    expect(buildSmsPhoneNumber('VN', '---')).toBe('');
  });

  it('rejects redacted / masked phone strings (breaks backend carrier lookup)', () => {
    expect(buildSmsPhoneNumber('VN', '+84*****9999')).toBe('');
    expect(buildSmsPhoneNumber('VN', '09**999999')).toBe('');
  });

  it('returns normalized E.164 for + with only digits, or + with spaces', () => {
    expect(buildSmsPhoneNumber('VN', '+84901234567')).toBe('+84901234567');
    expect(buildSmsPhoneNumber('TH', '+66812345678')).toBe('+66812345678');
    expect(buildSmsPhoneNumber('VN', '+84 901 234 567')).toBe('+84901234567');
  });

  it('prepends +84 for Vietnam national digits (strips one leading 0)', () => {
    expect(buildSmsPhoneNumber('VN', '0901234567')).toBe('+84901234567');
    expect(buildSmsPhoneNumber('VN', '901234567')).toBe('+84901234567');
    // Matches dev mock MOCK_SMS_SUCCESS_PHONES default (+84999999999)
    expect(buildSmsPhoneNumber('VN', '0999999999')).toBe('+84999999999');
  });

  it('prepends correct prefix for each SEA country', () => {
    expect(buildSmsPhoneNumber('TH', '0812345678')).toBe('+66812345678');
    expect(buildSmsPhoneNumber('SG', '91234567')).toBe('+6591234567');
    expect(buildSmsPhoneNumber('PH', '09171234567')).toBe('+639171234567');
  });

  it('strips non-digits from national input before prefixing', () => {
    expect(buildSmsPhoneNumber('VN', '(090) 123-4567')).toBe('+84901234567');
  });

  it('returns trimmed raw when ISO country has no dialing prefix mapping', () => {
    expect(buildSmsPhoneNumber('XX', 'anything')).toBe('anything');
  });
});

describe('getE164PlaceholderForCountry', () => {
  it('returns placeholder for each SEA country', () => {
    expect(getE164PlaceholderForCountry('VN')).toBe('+84901234567');
    expect(getE164PlaceholderForCountry('TH')).toBe('+66812345678');
    expect(getE164PlaceholderForCountry('SG')).toBe('+6591234567');
    expect(getE164PlaceholderForCountry('PH')).toBe('+639171234567');
  });
});

describe('validatePhoneNumber', () => {
  it('returns error when input is empty or whitespace only', () => {
    expect(validatePhoneNumber('VN', '')).toBe('Phone number is required');
    expect(validatePhoneNumber('VN', '   ')).toBe('Phone number is required');
  });

  it('returns error when input contains masking characters', () => {
    expect(validatePhoneNumber('VN', '+84*****9999')).toBe(
      'Masked phone numbers cannot be validated',
    );
    expect(validatePhoneNumber('VN', '09**999999')).toBe(
      'Masked phone numbers cannot be validated',
    );
  });

  it('accepts Vietnam national 10 digits with leading 0 or 9 digits without', () => {
    expect(validatePhoneNumber('VN', '0901234567')).toBeNull();
    expect(validatePhoneNumber('VN', '901234567')).toBeNull();
    expect(validatePhoneNumber('VN', '(090) 123-4567')).toBeNull();
  });

  it('accepts Vietnam E.164 +84 followed by 9 subscriber digits', () => {
    expect(validatePhoneNumber('VN', '+84901234567')).toBeNull();
    expect(validatePhoneNumber('VN', '+84 901 234 567')).toBeNull();
  });

  it('rejects Vietnam numbers with wrong length or prefix', () => {
    expect(validatePhoneNumber('VN', '090123456')).not.toBeNull();
    expect(validatePhoneNumber('VN', '+8490123456')).not.toBeNull();
    expect(validatePhoneNumber('VN', '+66812345678')).not.toBeNull();
  });

  it('accepts Thailand national and E.164 formats', () => {
    expect(validatePhoneNumber('TH', '0812345678')).toBeNull();
    expect(validatePhoneNumber('TH', '812345678')).toBeNull();
    expect(validatePhoneNumber('TH', '+66812345678')).toBeNull();
  });

  it('rejects Thailand numbers with wrong length', () => {
    expect(validatePhoneNumber('TH', '081234567')).not.toBeNull();
    expect(validatePhoneNumber('TH', '+6681234567')).not.toBeNull();
  });

  it('accepts Singapore 8-digit national and +65 E.164', () => {
    expect(validatePhoneNumber('SG', '91234567')).toBeNull();
    expect(validatePhoneNumber('SG', '+65 9123 4567')).toBeNull();
  });

  it('rejects Singapore numbers with leading 0 or wrong length', () => {
    expect(validatePhoneNumber('SG', '091234567')).not.toBeNull();
    expect(validatePhoneNumber('SG', '9123456')).not.toBeNull();
  });

  it('accepts Philippines national and E.164 formats', () => {
    expect(validatePhoneNumber('PH', '09171234567')).toBeNull();
    expect(validatePhoneNumber('PH', '9171234567')).toBeNull();
    expect(validatePhoneNumber('PH', '+63 917 123 4567')).toBeNull();
  });

  it('rejects Philippines numbers with wrong length', () => {
    expect(validatePhoneNumber('PH', '0917123456')).not.toBeNull();
    expect(validatePhoneNumber('PH', '+63917123456')).not.toBeNull();
  });

  it('returns error for unsupported country codes', () => {
    expect(validatePhoneNumber('XX', '+1234567890')).toBe(
      'Unsupported country code',
    );
  });
});

