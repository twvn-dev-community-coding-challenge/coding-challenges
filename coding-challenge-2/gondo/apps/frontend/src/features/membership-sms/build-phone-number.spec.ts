import { describe, it, expect } from 'vitest';

import {
  buildSmsPhoneNumber,
  getE164PlaceholderForCountry,
} from './build-phone-number';

describe('buildSmsPhoneNumber', () => {
  it('returns trimmed input when it already starts with + (E.164)', () => {
    expect(buildSmsPhoneNumber('VN', '+84901234567')).toBe('+84901234567');
    expect(buildSmsPhoneNumber('TH', '+66812345678')).toBe('+66812345678');
  });

  it('prepends +84 for Vietnam national digits (strips one leading 0)', () => {
    expect(buildSmsPhoneNumber('VN', '0901234567')).toBe('+84901234567');
    expect(buildSmsPhoneNumber('VN', '901234567')).toBe('+84901234567');
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
