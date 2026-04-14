import { screen, fireEvent, waitFor } from '@testing-library/react';
import { MemoryRouter } from 'react-router';
import { describe, it, expect, vi, beforeEach } from 'vitest';

import { renderWithTheme } from '../../test-utils';
import { MembershipSmsFeature } from './membership-sms-feature';

const mockCreateNotification = vi.fn();
const mockDispatchNotification = vi.fn();
const mockGetNotification = vi.fn();
const mockListNotifications = vi.fn();
const mockRetryNotification = vi.fn();

vi.mock('@gondo/ts-core', () => ({
  createNotificationApi: () => ({
    createNotification: mockCreateNotification,
    dispatchNotification: mockDispatchNotification,
    getNotification: mockGetNotification,
    listNotifications: mockListNotifications,
    retryNotification: mockRetryNotification,
  }),
  generateSixDigitOtp: () => '123456',
}));

const successNotification = {
  notification_id: 'notif-1',
  state: 'New',
  message_id: 'msg-1',
  channel_type: 'SMS',
  recipient: 'test',
  content: 'test',
  channel_payload: {},
  attempt: 0,
  selected_provider_id: null,
  routing_rule_version: null,
  created_at: '2026-04-11T12:00:00Z',
  updated_at: '2026-04-11T12:00:00Z',
} as const;

const queuedNotification = {
  notification_id: 'notif-1',
  state: 'Queue',
  message_id: 'msg-1',
  channel_type: 'SMS',
  recipient: 'test',
  content: 'test',
  channel_payload: {},
  attempt: 1,
  selected_provider_id: 'twilio',
  routing_rule_version: 1,
  created_at: '2026-04-11T12:00:00Z',
  updated_at: '2026-04-11T12:00:01Z',
} as const;

const renderFeature = () =>
  renderWithTheme(
    <MemoryRouter>
      <MembershipSmsFeature />
    </MemoryRouter>,
  );

describe('MembershipSmsFeature', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('renders the form in idle state', () => {
    renderFeature();
    expect(screen.getByRole('button', { name: /send sms/i })).toBeTruthy();
    expect(screen.getByLabelText(/message id/i)).toBeTruthy();
  });

  it('submits the form: creates notification then dispatches', async () => {
    mockCreateNotification.mockResolvedValue({
      ok: true,
      value: { ...successNotification },
    });
    mockDispatchNotification.mockResolvedValue({
      ok: true,
      value: { ...queuedNotification },
    });

    renderFeature();

    fireEvent.change(screen.getByLabelText(/phone number/i), {
      target: { value: '+84901234567' },
    });

    fireEvent.click(screen.getByRole('button', { name: /send sms/i }));

    await waitFor(() => {
      expect(screen.getByText('123456')).toBeTruthy();
    });

    expect(mockCreateNotification).toHaveBeenCalledTimes(1);
    expect(mockCreateNotification).toHaveBeenCalledWith({
      message_id: expect.any(String),
      recipient: '',
      content: expect.stringContaining('123456'),
      channel_payload: {
        country_code: 'VN',
        phone_number: '+84901234567',
      },
    });
    expect(mockCreateNotification.mock.calls[0]?.[0]).not.toHaveProperty('channel_type');
    expect(mockDispatchNotification).toHaveBeenCalledTimes(1);
  });

  it('builds E.164 phone_number from national digits before create', async () => {
    mockCreateNotification.mockResolvedValue({
      ok: true,
      value: { ...successNotification },
    });
    mockDispatchNotification.mockResolvedValue({
      ok: true,
      value: { ...queuedNotification },
    });

    renderFeature();

    fireEvent.change(screen.getByLabelText(/phone number/i), {
      target: { value: '0901234567' },
    });

    fireEvent.click(screen.getByRole('button', { name: /send sms/i }));

    await waitFor(() => {
      expect(mockCreateNotification).toHaveBeenCalled();
    });

    expect(mockCreateNotification).toHaveBeenCalledWith(
      expect.objectContaining({
        channel_payload: {
          country_code: 'VN',
          phone_number: '+84901234567',
        },
      }),
    );
  });

  it('shows error message when create notification fails', async () => {
    mockCreateNotification.mockResolvedValue({
      ok: false,
      error: { code: 'VALIDATION_ERROR', message: 'message_id already exists' },
    });

    renderFeature();
    fireEvent.click(screen.getByRole('button', { name: /send sms/i }));

    await waitFor(() => {
      expect(screen.getByRole('alert')).toBeTruthy();
    });
  });

  it('shows success alert with link to notification tracking after dispatch', async () => {
    mockCreateNotification.mockResolvedValue({
      ok: true,
      value: { ...successNotification },
    });
    mockDispatchNotification.mockResolvedValue({
      ok: true,
      value: { ...queuedNotification },
    });

    renderFeature();
    fireEvent.click(screen.getByRole('button', { name: /send sms/i }));

    await waitFor(() => {
      expect(
        screen.getByRole('link', { name: /notification tracking/i }),
      ).toBeTruthy();
    });

    const link = screen.getByRole('link', { name: /notification tracking/i });
    expect(link.getAttribute('href')).toBe('/tracking');
  });

  it('does not show countdown timer on membership page after send', async () => {
    mockCreateNotification.mockResolvedValue({
      ok: true,
      value: { ...successNotification },
    });
    mockDispatchNotification.mockResolvedValue({
      ok: true,
      value: { ...queuedNotification },
    });

    renderFeature();
    fireEvent.click(screen.getByRole('button', { name: /send sms/i }));

    await waitFor(() => {
      expect(screen.getByText('123456')).toBeTruthy();
    });

    expect(screen.queryByRole('timer')).toBeNull();
  });

  it('allows sending a new SMS after clicking New SMS button', async () => {
    mockCreateNotification.mockResolvedValue({
      ok: true,
      value: { ...successNotification },
    });
    mockDispatchNotification.mockResolvedValue({
      ok: true,
      value: { ...queuedNotification },
    });

    renderFeature();
    fireEvent.click(screen.getByRole('button', { name: /send sms/i }));

    await waitFor(() => {
      expect(screen.getByText('123456')).toBeTruthy();
    });

    fireEvent.click(screen.getByRole('button', { name: /new sms/i }));

    expect(screen.getByRole('button', { name: /send sms/i })).toBeTruthy();
  });
});
