import { screen, waitFor } from '@testing-library/react';
import { MemoryRouter } from 'react-router';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import { renderWithTheme } from '../test-utils';
import { AppRoutes } from './app';

const mockListNotifications = vi.fn();

vi.mock('@gondo/ts-core', () => ({
  createNotificationApi: () => ({
    createNotification: vi.fn(),
    dispatchNotification: vi.fn(),
    getNotification: vi.fn(),
    getSmsKpis: vi.fn().mockResolvedValue({
      ok: true,
      value: {
        source: 'in_memory_notifications',
        created_at_filter: { from: null, to: null },
        currency_note: '',
        overall: {
          volume: 0,
          total_notifications: 0,
          total_estimated_cost: 0,
          total_actual_cost: 0,
          with_estimate_count: 0,
          with_actual_count: 0,
          send_success: 0,
          send_failed: 0,
          carrier_rejected: 0,
          in_flight: 0,
          terminal_failure: 0,
          terminal_success_rate: null,
          terminal_failure_rate: null,
        },
        by_provider: [],
        by_country: [],
        by_calling_domain: [],
      },
    }),
    listNotifications: mockListNotifications,
    retryNotification: vi.fn(),
  }),
  generateSixDigitOtp: () => '123456',
}));

describe('App', () => {
  beforeEach(() => {
    mockListNotifications.mockResolvedValue({ ok: true, value: [] });
  });
  it('renders the NavBar with navigation', () => {
    renderWithTheme(
      <MemoryRouter initialEntries={['/membership']}>
        <AppRoutes />
      </MemoryRouter>,
    );

    expect(screen.getByRole('navigation', { name: /main/i })).toBeTruthy();
    expect(screen.getByText('Gondo SMS Platform')).toBeTruthy();
  });

  it('renders the Membership Registration page heading at /membership route', () => {
    renderWithTheme(
      <MemoryRouter initialEntries={['/membership']}>
        <AppRoutes />
      </MemoryRouter>,
    );

    expect(
      screen.getByRole('heading', { name: /membership registration/i }),
    ).toBeTruthy();
  });

  it('renders the Notification Tracking page heading at /tracking route', async () => {
    renderWithTheme(
      <MemoryRouter initialEntries={['/tracking']}>
        <AppRoutes />
      </MemoryRouter>,
    );

    expect(
      await screen.findByRole('heading', { name: /notification tracking/i }),
    ).toBeTruthy();
    await waitFor(() => expect(mockListNotifications).toHaveBeenCalled());
  });
});
