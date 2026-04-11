import { screen } from '@testing-library/react';

import { renderWithTheme } from '../../test-utils';
import { NotificationStatusBadge } from './notification-status-badge';

const expectChipHasColorClass = (
  label: string,
  colorFragment: string,
): void => {
  const chip = screen.getByText(label).closest('.MuiChip-root');
  expect(chip).toBeTruthy();
  expect(chip?.className.includes(colorFragment)).toBe(true);
};

describe('NotificationStatusBadge', () => {
  it('renders the status label text', () => {
    renderWithTheme(<NotificationStatusBadge statusLabel="Queue" />);
    expect(screen.getByText('Queue')).toBeTruthy();
  });

  it('applies success styling for Send-success', () => {
    renderWithTheme(<NotificationStatusBadge statusLabel="Send-success" />);
    expectChipHasColorClass('Send-success', 'MuiChip-colorSuccess');
  });

  it('applies error styling for Send-failed', () => {
    renderWithTheme(<NotificationStatusBadge statusLabel="Send-failed" />);
    expectChipHasColorClass('Send-failed', 'MuiChip-colorError');
  });

  it('applies warning styling for Carrier-rejected', () => {
    renderWithTheme(<NotificationStatusBadge statusLabel="Carrier-rejected" />);
    expectChipHasColorClass('Carrier-rejected', 'MuiChip-colorWarning');
  });

  it('applies default/neutral styling for other states', () => {
    renderWithTheme(<NotificationStatusBadge statusLabel="New" />);
    expectChipHasColorClass('New', 'MuiChip-colorDefault');
  });
});
