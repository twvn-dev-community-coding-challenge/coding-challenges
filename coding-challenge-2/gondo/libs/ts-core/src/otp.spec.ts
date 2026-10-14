import { describe, expect, it } from 'vitest';
import { generateSixDigitOtp } from './otp';

describe('generateSixDigitOtp', () => {
  it('generates a string of exactly length 6', () => {
    const otp = generateSixDigitOtp();
    expect(otp).toHaveLength(6);
  });

  it('contains only digits 0-9', () => {
    const otp = generateSixDigitOtp();
    expect(otp).toMatch(/^[0-9]{6}$/);
  });

  it('generates different values on multiple calls (probabilistic - run 10 times, at least 2 unique)', () => {
    const values = new Set<string>();
    for (let i = 0; i < 10; i += 1) {
      values.add(generateSixDigitOtp());
    }
    expect(values.size).toBeGreaterThanOrEqual(2);
  });
});
