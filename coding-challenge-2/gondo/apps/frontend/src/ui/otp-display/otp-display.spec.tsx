import { screen } from '@testing-library/react';

import { renderWithTheme } from '../../test-utils';
import { OtpDisplay } from './otp-display';

describe('OtpDisplay', () => {
  it('renders the OTP code visibly', () => {
    renderWithTheme(<OtpDisplay otp="123456" />);
    expect(screen.getByText('123456')).toBeTruthy();
  });

  it('has accessible label', () => {
    renderWithTheme(<OtpDisplay otp="654321" />);
    const status = screen.getByRole('status', { name: /OTP/i });
    expect(status).toBeTruthy();
  });

  it('renders each digit separately for styling', () => {
    renderWithTheme(<OtpDisplay otp="012345" />);
    const status = screen.getByRole('status');
    const digitRow = status.querySelector('[aria-hidden="true"]');
    const spans = digitRow?.querySelectorAll('span') ?? [];
    expect(spans).toHaveLength(6);
    expect(spans[0]?.textContent).toBe('0');
    expect(spans[1]?.textContent).toBe('1');
    expect(spans[2]?.textContent).toBe('2');
    expect(spans[3]?.textContent).toBe('3');
    expect(spans[4]?.textContent).toBe('4');
    expect(spans[5]?.textContent).toBe('5');
  });
});
