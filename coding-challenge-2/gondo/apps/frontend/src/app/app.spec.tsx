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
