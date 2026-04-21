import {
  act,
  fireEvent,
  screen,
  waitFor,
  within,
} from '@testing-library/react';
import { useEffect, type ReactNode } from 'react';
import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';

import type { CountdownEntry } from '../../context';
import { useCountdown } from '../../context';
import { renderWithTheme } from '../../test-utils';
import { NotificationTrackingPage } from './notification-tracking-page';

const mockListNotifications = vi.fn();
const mockRetryNotification = vi.fn();
const mockGetPipelineEvents = vi.fn();

vi.mock('@gondo/ts-core', () => ({
  createNotificationApi: () => ({
    listNotifications: mockListNotifications,
    createNotification: vi.fn(),
    dispatchNotification: vi.fn(),
    getNotification: vi.fn(),
    getPipelineEvents: mockGetPipelineEvents,
    retryNotification: mockRetryNotification,
  }),
  generateSixDigitOtp: vi.fn(() => '654321'),
}));

const CREATED_AT_MS = new Date('2026-04-11T12:30:00Z').getTime();

const notificationA = {
  notification_id: 'abc-123',
  message_id: 'msg-new',
  channel_type: 'SMS',
  recipient: 'user@example.com',
  content: 'Your OTP is 123456',
  channel_payload: { country_code: 'VN', phone_number: '+84901234567' },
  state: 'Queue',
  attempt: 1,
  selected_provider_id: 'prv_02',
  routing_rule_version: 1,
  created_at: '2026-04-11T12:30:00Z',
  updated_at: '2026-04-11T12:30:01Z',
} as const;

const notificationB = {
  notification_id: 'def-456',
  message_id: 'msg-old',
  channel_type: 'SMS',
  recipient: 'other@example.com',
  content: 'Hi',
  channel_payload: { country_code: 'US', phone_number: '+15551234567' },
  state: 'New',
  attempt: 0,
  selected_provider_id: null,
  routing_rule_version: null,
  created_at: '2026-04-11T10:30:00Z',
  updated_at: '2026-04-11T10:30:00Z',
} as const;

const notificationSendSuccess = {
  ...notificationA,
  notification_id: 'succ-789',
  message_id: 'msg-send-success',
  state: 'Send-success',
} as const;

const notificationSendFailed = {
  ...notificationA,
  notification_id: 'fail-321',
  message_id: 'msg-send-failed',
  state: 'Send-failed',
} as const;

const notificationNewExpired = {
  ...notificationA,
  notification_id: 'new-exp-1',
  message_id: 'msg-new-expired',
  state: 'New',
  attempt: 0,
} as const;

const retryButtonNameForState = (state: string): RegExp =>
  new RegExp(`retry notification \\(state: ${state}\\)`, 'i');

const SeedCountdownEntry = ({
  entry,
  children,
}: {
  readonly entry: CountdownEntry;
  readonly children: ReactNode;
}) => {
  const { startCountdown } = useCountdown();
  useEffect(() => {
    startCountdown(entry);
  }, [entry, startCountdown]);
  return children;
};

describe('NotificationTrackingPage', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  afterEach(() => {
    vi.useRealTimers();
  });

  it('renders the page heading', async () => {
    mockListNotifications.mockResolvedValue({ ok: true, value: [] });

    renderWithTheme(<NotificationTrackingPage />);

    expect(
      screen.getByRole('heading', { name: /notification tracking/i }),
    ).toBeTruthy();
    await waitFor(() => expect(mockListNotifications).toHaveBeenCalledTimes(1));
  });

  it('calls listNotifications on mount', async () => {
    mockListNotifications.mockResolvedValue({ ok: true, value: [] });

    renderWithTheme(<NotificationTrackingPage />);

    await waitFor(() => expect(mockListNotifications).toHaveBeenCalledTimes(1));
  });

  it('shows a loading indicator while fetching', async () => {
    let resolveList: ((value: unknown) => void) | undefined;
    mockListNotifications.mockImplementation(
      () =>
        new Promise((resolve) => {
          resolveList = resolve;
        }),
    );

    renderWithTheme(<NotificationTrackingPage />);

    expect(screen.getByRole('progressbar')).toBeTruthy();

    await act(async () => {
      resolveList?.({ ok: true, value: [] });
    });

    await waitFor(() => {
      expect(screen.queryByRole('progressbar')).toBeNull();
    });
  });

  it('renders table rows for notifications (newest first)', async () => {
    mockListNotifications.mockResolvedValue({
      ok: true,
      value: [notificationB, notificationA],
    });

    renderWithTheme(<NotificationTrackingPage />);

    await waitFor(() => {
      expect(screen.getByText('msg-new')).toBeTruthy();
    });

    const rows = screen.getAllByRole('row');
    expect(rows.length).toBeGreaterThanOrEqual(3);
    const bodyRows = screen.getAllByRole('row').slice(1);
    expect(bodyRows[0]?.textContent ?? '').toContain('msg-new');
    expect(bodyRows[1]?.textContent ?? '').toContain('msg-old');
  });

  it('shows an alert when fetch fails', async () => {
    mockListNotifications.mockResolvedValue({
      ok: false,
      error: { code: 'HTTP_ERROR', message: 'Request failed' },
    });

    renderWithTheme(<NotificationTrackingPage />);

    expect(await screen.findByText(/failed to load notifications/i)).toBeTruthy();
  });

  it('re-fetches when Refresh is clicked', async () => {
    mockListNotifications.mockResolvedValue({ ok: true, value: [] });

    renderWithTheme(<NotificationTrackingPage />);

    await waitFor(() => expect(mockListNotifications).toHaveBeenCalledTimes(1));

    fireEvent.click(screen.getByRole('button', { name: /refresh/i }));

    await waitFor(() => expect(mockListNotifications).toHaveBeenCalledTimes(2));
  });

  it('shows countdown timer from created_at without context entry (survives refresh)', async () => {
    vi.spyOn(Date, 'now').mockReturnValue(CREATED_AT_MS + 30_000);
    mockListNotifications.mockResolvedValue({
      ok: true,
      value: [notificationA],
    });

    renderWithTheme(<NotificationTrackingPage />);

    await waitFor(() => {
      expect(screen.getByRole('timer')).toBeTruthy();
    });

    expect(screen.getByRole('columnheader', { name: /actions/i })).toBeTruthy();
  });

  it('shows Retry when created_at countdown has expired without context entry', async () => {
    vi.spyOn(Date, 'now').mockReturnValue(CREATED_AT_MS + 200_000);
    mockListNotifications.mockResolvedValue({
      ok: true,
      value: [notificationA],
    });

    renderWithTheme(<NotificationTrackingPage />);

    expect(
      await screen.findByRole('button', {
        name: retryButtonNameForState('Queue'),
      }),
    ).toBeTruthy();
  });

  it('shows countdown timer and OTP when context entry exists', async () => {
    vi.spyOn(Date, 'now').mockReturnValue(CREATED_AT_MS);
    mockListNotifications.mockResolvedValue({
      ok: true,
      value: [notificationA],
    });

    const entry: CountdownEntry = {
      notificationId: notificationA.notification_id,
      otp: '999888',
      startedAt: CREATED_AT_MS,
      durationSeconds: 120,
    };

    renderWithTheme(
      <SeedCountdownEntry entry={entry}>
        <NotificationTrackingPage />
      </SeedCountdownEntry>,
    );

    await waitFor(() => {
      expect(screen.getByRole('timer')).toBeTruthy();
    });

    expect(screen.getByRole('status', { name: /otp code/i })).toBeTruthy();
  });

  it('uses context startedAt when more recent than created_at (retry case)', async () => {
    const retryStartedAt = CREATED_AT_MS + 150_000;
    vi.spyOn(Date, 'now').mockReturnValue(retryStartedAt + 10_000);
    mockListNotifications.mockResolvedValue({
      ok: true,
      value: [notificationA],
    });

    const entry: CountdownEntry = {
      notificationId: notificationA.notification_id,
      otp: '111222',
      startedAt: retryStartedAt,
      durationSeconds: 120,
    };

    renderWithTheme(
      <SeedCountdownEntry entry={entry}>
        <NotificationTrackingPage />
      </SeedCountdownEntry>,
    );

    await waitFor(() => {
      expect(screen.getByRole('timer')).toBeTruthy();
    });
  });

  it('calls retryNotification and re-fetches the list when Retry succeeds', async () => {
    vi.spyOn(Date, 'now').mockReturnValue(CREATED_AT_MS + 200_000);
    mockListNotifications.mockResolvedValue({
      ok: true,
      value: [notificationA],
    });
    mockRetryNotification.mockResolvedValue({
      ok: true,
      value: { ...notificationA, state: 'Queue', attempt: 2 },
    });

    renderWithTheme(<NotificationTrackingPage />);

    const retryButton = await screen.findByRole('button', {
      name: retryButtonNameForState('Queue'),
    });

    await act(async () => {
      fireEvent.click(retryButton);
    });

    await waitFor(() => {
      expect(mockRetryNotification).toHaveBeenCalledWith(notificationA.notification_id, {
        as_of: expect.any(String),
      });
    });

    await waitFor(() => {
      expect(mockListNotifications.mock.calls.length).toBeGreaterThanOrEqual(2);
    });
  });

  it('opens pipeline logs dialog and shows runtime events JSON', async () => {
    mockListNotifications.mockResolvedValue({
      ok: true,
      value: [notificationA],
    });
    mockGetPipelineEvents.mockResolvedValue({
      ok: true,
      value: {
        notification_id: notificationA.notification_id,
        message_id: notificationA.message_id,
        events: [
          {
            timestamp: '2026-04-11T12:30:05.000000+00:00',
            phase: 'dispatch.begin',
            detail: { carrier: 'VN' },
          },
        ],
      },
    });

    renderWithTheme(<NotificationTrackingPage />);

    await waitFor(() => {
      expect(screen.getByText('msg-new')).toBeTruthy();
    });

    fireEvent.click(
      screen.getByRole('button', { name: /^pipeline logs$/i }),
    );

    await waitFor(() => {
      expect(mockGetPipelineEvents).toHaveBeenCalledWith(notificationA.notification_id);
    });

    expect(
      await screen.findByRole('heading', {
        name: /sms pipeline \(runtime\)/i,
      }),
    ).toBeTruthy();

    const dialog = screen.getByRole('dialog');
    expect(within(dialog).getByText(/^begin$/)).toBeTruthy();
    expect(within(dialog).getByText('VN')).toBeTruthy();
  });

  it('shows Retry after expiry for Send-success (terminal state)', async () => {
    vi.spyOn(Date, 'now').mockReturnValue(CREATED_AT_MS + 200_000);
    mockListNotifications.mockResolvedValue({
      ok: true,
      value: [notificationSendSuccess],
    });

    renderWithTheme(<NotificationTrackingPage />);

    expect(
      await screen.findByRole('button', {
        name: retryButtonNameForState('Send-success'),
      }),
    ).toBeTruthy();
  });

  it('shows inline error when retry is rejected for Send-success (409)', async () => {
    vi.spyOn(Date, 'now').mockReturnValue(CREATED_AT_MS + 200_000);
    mockListNotifications.mockResolvedValue({
      ok: true,
      value: [notificationSendSuccess],
    });
    const conflictMessage = 'Notification is not in a retryable state';
    mockRetryNotification.mockResolvedValue({
      ok: false,
      error: { code: 'CONFLICT', message: conflictMessage },
    });

    renderWithTheme(<NotificationTrackingPage />);

    const retryButton = await screen.findByRole('button', {
      name: retryButtonNameForState('Send-success'),
    });

    await act(async () => {
      fireEvent.click(retryButton);
    });

    const alert = await screen.findByRole('alert');
    expect(alert.textContent?.includes(conflictMessage)).toBe(true);
  });

  it('calls retry and re-fetches when Send-failed is retryable', async () => {
    vi.spyOn(Date, 'now').mockReturnValue(CREATED_AT_MS + 200_000);
    mockListNotifications.mockResolvedValue({
      ok: true,
      value: [notificationSendFailed],
    });
    mockRetryNotification.mockResolvedValue({
      ok: true,
      value: { ...notificationSendFailed, state: 'Send-to-provider', attempt: 2 },
    });

    renderWithTheme(<NotificationTrackingPage />);

    const retryButton = await screen.findByRole('button', {
      name: retryButtonNameForState('Send-failed'),
    });

    await act(async () => {
      fireEvent.click(retryButton);
    });

    await waitFor(() => {
      expect(mockRetryNotification).toHaveBeenCalledWith(
        notificationSendFailed.notification_id,
        { as_of: expect.any(String) },
      );
    });

    await waitFor(() => {
      expect(mockListNotifications.mock.calls.length).toBeGreaterThanOrEqual(2);
    });
  });

  it('shows inline error when retry is rejected for New state', async () => {
    vi.spyOn(Date, 'now').mockReturnValue(CREATED_AT_MS + 200_000);
    mockListNotifications.mockResolvedValue({
      ok: true,
      value: [notificationNewExpired],
    });
    const conflictMessage = 'Notification is not in a retryable state';
    mockRetryNotification.mockResolvedValue({
      ok: false,
      error: { code: 'CONFLICT', message: conflictMessage },
    });

    renderWithTheme(<NotificationTrackingPage />);

    const retryButton = await screen.findByRole('button', {
      name: retryButtonNameForState('New'),
    });

    await act(async () => {
      fireEvent.click(retryButton);
    });

    const alert = await screen.findByRole('alert');
    expect(alert.textContent?.includes(conflictMessage)).toBe(true);
  });

  it('keeps Retry clickable after a 409 so the user can try again', async () => {
    vi.spyOn(Date, 'now').mockReturnValue(CREATED_AT_MS + 200_000);
    mockListNotifications.mockResolvedValue({
      ok: true,
      value: [notificationSendSuccess],
    });
    mockRetryNotification.mockResolvedValue({
      ok: false,
      error: {
        code: 'CONFLICT',
        message: 'Notification is not in a retryable state',
      },
    });

    renderWithTheme(<NotificationTrackingPage />);

    const retryButton = await screen.findByRole('button', {
      name: retryButtonNameForState('Send-success'),
    });

    await act(async () => {
      fireEvent.click(retryButton);
    });

    await screen.findByRole('alert');

    expect(retryButton.hasAttribute('disabled')).toBe(false);

    await act(async () => {
      fireEvent.click(retryButton);
    });

    await waitFor(() => {
      expect(mockRetryNotification).toHaveBeenCalledTimes(2);
    });
  });
});
