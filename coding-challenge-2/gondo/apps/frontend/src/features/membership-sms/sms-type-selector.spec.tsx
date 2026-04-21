import { screen, fireEvent } from '@testing-library/react';
import { describe, it, expect, vi } from 'vitest';

import { renderWithTheme } from '../../test-utils';
import { SmsTypeSelector } from './sms-type-selector';

describe('SmsTypeSelector', () => {
  it('renders "Registration Success + OTP" option as selected radio/select', () => {
    renderWithTheme(
      <SmsTypeSelector
        value="registration-otp"
        onChange={vi.fn()}
      />,
    );

    const select = screen.getByLabelText(/sms type/i) as HTMLSelectElement;
    expect(select.value).toBe('registration-otp');
    expect(
      screen.getByRole('option', { name: /registration success \+ otp/i }),
    ).toBeTruthy();
  });

  it('calls onChange when a different scenario is selected (future-proof, but only one option now)', () => {
    const onChange = vi.fn();
    renderWithTheme(
      <SmsTypeSelector
        value="registration-otp"
        onChange={onChange}
      />,
    );

    const select = screen.getByLabelText(/sms type/i);
    fireEvent.change(select, { target: { value: 'registration-otp' } });

    expect(onChange).toHaveBeenCalledWith('registration-otp');
  });

  it('renders with accessible label "SMS Type"', () => {
    renderWithTheme(
      <SmsTypeSelector
        value="registration-otp"
        onChange={vi.fn()}
      />,
    );

    expect(screen.getAllByText('SMS Type').length).toBeGreaterThan(0);
  });
});
