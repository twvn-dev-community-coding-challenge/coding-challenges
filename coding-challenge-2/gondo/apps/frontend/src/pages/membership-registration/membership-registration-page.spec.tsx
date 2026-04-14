import { screen } from '@testing-library/react';
import { describe, it, expect, vi } from 'vitest';

import { renderWithTheme } from '../../test-utils';
import { MembershipRegistrationPage } from './membership-registration-page';

vi.mock('../../features/membership-sms', () => ({
  MembershipSmsFeature: () => (
    <div data-testid="membership-sms-feature">MembershipSmsFeature</div>
  ),
}));

describe('MembershipRegistrationPage', () => {
  it('renders the page heading "Membership Registration"', () => {
    renderWithTheme(<MembershipRegistrationPage />);

    expect(
      screen.getByRole('heading', { name: /membership registration/i }),
    ).toBeTruthy();
  });

  it('renders the MembershipSmsFeature component', () => {
    renderWithTheme(<MembershipRegistrationPage />);

    expect(screen.getByTestId('membership-sms-feature')).toBeTruthy();
  });
});
