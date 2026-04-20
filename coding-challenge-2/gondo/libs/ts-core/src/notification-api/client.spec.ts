import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { createNotificationApi } from './client';
import type {
  CreateNotificationRequest,
  NotificationResource,
  PipelineEventsData,
  SmsKpisData,
} from './types';

const sampleNotification: NotificationResource = {
  notification_id: 'abc',
  message_id: 'msg-1',
  channel_type: 'SMS',
  recipient: '+15551234567',
  content: 'Hello',
  channel_payload: { country_code: 'US', phone_number: '+15551234567' },
  state: 'queued',
  attempt: 0,
  selected_provider_id: null,
  routing_rule_version: 1,
  created_at: '2026-04-11T00:00:00Z',
  updated_at: '2026-04-11T00:00:00Z',
};

/** Matches OpenAPI: channel_type is optional (server default SMS). */
const minimalCreateRequest: CreateNotificationRequest = {
  message_id: 'msg-1',
  recipient: '+15551234567',
  content: 'Hello',
  channel_payload: { country_code: 'US', phone_number: '+15551234567' },
};

describe('notification API client', () => {
  beforeEach(() => {
    vi.stubGlobal('fetch', vi.fn());
  });

  afterEach(() => {
    vi.unstubAllGlobals();
  });

  describe('createNotification', () => {
    it('successful 201 response', async () => {
      const mockFetch = vi.mocked(globalThis.fetch);
      mockFetch.mockResolvedValue({
        ok: true,
        status: 201,
        statusText: 'Created',
        json: () => ({ data: sampleNotification }),
      } as Response);

      const api = createNotificationApi();
      const result = await api.createNotification(minimalCreateRequest);

      expect(result).toEqual({ ok: true, value: sampleNotification });
      expect(mockFetch).toHaveBeenCalledTimes(1);
      const [url, init] = mockFetch.mock.calls[0] as [string, RequestInit];
      expect(url).toBe('http://localhost:8001/notifications');
      expect(init?.method).toBe('POST');
      expect(init?.headers).toMatchObject({
        'Content-Type': 'application/json',
        Accept: 'application/json',
      });
      expect(JSON.parse(String(init?.body))).toEqual(minimalCreateRequest);
    });

    it('includes channel_type in JSON body when provided', async () => {
      const mockFetch = vi.mocked(globalThis.fetch);
      mockFetch.mockResolvedValue({
        ok: true,
        status: 201,
        statusText: 'Created',
        json: () => ({ data: sampleNotification }),
      } as Response);

      const withChannel: CreateNotificationRequest = {
        ...minimalCreateRequest,
        channel_type: 'SMS',
      };

      const api = createNotificationApi();
      await api.createNotification(withChannel);

      const [, init] = mockFetch.mock.calls[0] as [string, RequestInit];
      expect(JSON.parse(String(init?.body))).toEqual(withChannel);
    });

    it('API error response (422)', async () => {
      const mockFetch = vi.mocked(globalThis.fetch);
      mockFetch.mockResolvedValue({
        ok: false,
        status: 422,
        statusText: 'Unprocessable Entity',
        json: () => ({
          error: { code: 'VALIDATION_ERROR', message: 'bad input' },
        }),
      } as Response);

      const api = createNotificationApi();
      const result = await api.createNotification(minimalCreateRequest);

      expect(result).toEqual({
        ok: false,
        error: { code: 'VALIDATION_ERROR', message: 'bad input' },
      });
    });

    it('network error (fetch throws)', async () => {
      const mockFetch = vi.mocked(globalThis.fetch);
      mockFetch.mockRejectedValue(new TypeError('Failed to fetch'));

      const api = createNotificationApi();
      const result = await api.createNotification(minimalCreateRequest);

      expect(result.ok).toBe(false);
      if (!result.ok) {
        expect(result.error.code).toBe('NETWORK_ERROR');
        expect(result.error.message).toContain('Failed to fetch');
      }
    });
  });

  describe('dispatchNotification', () => {
    it('successful 200', async () => {
      const mockFetch = vi.mocked(globalThis.fetch);
      mockFetch.mockResolvedValue({
        ok: true,
        status: 200,
        statusText: 'OK',
        json: () => ({ data: sampleNotification }),
      } as Response);

      const api = createNotificationApi();
      const result = await api.dispatchNotification('notif-id', {
        as_of: '2026-04-11T00:00:00Z',
      });

      expect(result).toEqual({ ok: true, value: sampleNotification });
      expect(mockFetch).toHaveBeenCalledWith(
        'http://localhost:8001/notifications/notif-id/dispatch',
        expect.objectContaining({
          method: 'POST',
          body: JSON.stringify({ as_of: '2026-04-11T00:00:00Z' }),
        }),
      );
    });
  });

  describe('getNotification', () => {
    it('successful 200', async () => {
      const mockFetch = vi.mocked(globalThis.fetch);
      mockFetch.mockResolvedValue({
        ok: true,
        status: 200,
        statusText: 'OK',
        json: () => ({ data: sampleNotification }),
      } as Response);

      const api = createNotificationApi();
      const result = await api.getNotification('notif-id');

      expect(result).toEqual({ ok: true, value: sampleNotification });
      const [url, init] = mockFetch.mock.calls[0] as [string, RequestInit];
      expect(url).toBe('http://localhost:8001/notifications/notif-id');
      expect(init?.method).toBe('GET');
    });
  });

  describe('listNotifications', () => {
    it('successful 200 returns notifications array', async () => {
      const mockFetch = vi.mocked(globalThis.fetch);
      const listPayload = {
        data: {
          notifications: [sampleNotification],
        },
      };
      mockFetch.mockResolvedValue({
        ok: true,
        status: 200,
        statusText: 'OK',
        json: () => listPayload,
      } as Response);

      const api = createNotificationApi();
      const result = await api.listNotifications();

      expect(result).toEqual({ ok: true, value: [sampleNotification] });
      const [url, init] = mockFetch.mock.calls[0] as [string, RequestInit];
      expect(url).toBe('http://localhost:8001/notifications');
      expect(init?.method).toBe('GET');
    });

    it('successful 200 with empty notifications array', async () => {
      const mockFetch = vi.mocked(globalThis.fetch);
      mockFetch.mockResolvedValue({
        ok: true,
        status: 200,
        statusText: 'OK',
        json: () => ({ data: { notifications: [] } }),
      } as Response);

      const api = createNotificationApi();
      const result = await api.listNotifications();

      expect(result).toEqual({ ok: true, value: [] });
    });

    it('missing data.notifications yields INVALID_RESPONSE', async () => {
      const mockFetch = vi.mocked(globalThis.fetch);
      mockFetch.mockResolvedValue({
        ok: true,
        status: 200,
        statusText: 'OK',
        json: () => ({ data: {} }),
      } as Response);

      const api = createNotificationApi();
      const result = await api.listNotifications();

      expect(result).toEqual({
        ok: false,
        error: { code: 'INVALID_RESPONSE', message: 'Missing notifications' },
      });
    });
  });

  describe('retryNotification', () => {
    it('successful 200', async () => {
      const mockFetch = vi.mocked(globalThis.fetch);
      mockFetch.mockResolvedValue({
        ok: true,
        status: 200,
        statusText: 'OK',
        json: () => ({ data: sampleNotification }),
      } as Response);

      const api = createNotificationApi();
      const result = await api.retryNotification('notif-id', { as_of: null });

      expect(result).toEqual({ ok: true, value: sampleNotification });
      expect(mockFetch).toHaveBeenCalledWith(
        'http://localhost:8001/notifications/notif-id/retry',
        expect.objectContaining({
          method: 'POST',
          body: JSON.stringify({ as_of: null }),
        }),
      );
    });
  });

  describe('getSmsKpis', () => {
    const kpisPayload: SmsKpisData = {
      source: 'in_memory_notifications',
      currency_note: 'note',
      overall: {
        volume: 1,
        total_notifications: 1,
        total_estimated_cost: 0,
        total_actual_cost: 0,
        with_estimate_count: 0,
        with_actual_count: 0,
        send_success: 0,
        send_failed: 0,
        carrier_rejected: 0,
        in_flight: 1,
        terminal_failure: 0,
        terminal_success_rate: null,
        terminal_failure_rate: null,
      },
      by_provider: [],
      by_country: [],
    };

    it('successful 200', async () => {
      const mockFetch = vi.mocked(globalThis.fetch);
      mockFetch.mockResolvedValue({
        ok: true,
        status: 200,
        statusText: 'OK',
        json: () => ({ data: kpisPayload }),
      } as Response);

      const api = createNotificationApi();
      const result = await api.getSmsKpis();

      expect(result).toEqual({ ok: true, value: kpisPayload });
      expect(mockFetch).toHaveBeenCalledWith(
        'http://localhost:8001/notifications/kpis',
        expect.objectContaining({ method: 'GET' }),
      );
    });
  });

  describe('getPipelineEvents', () => {
    const pipelinePayload: PipelineEventsData = {
      notification_id: 'notif-id',
      message_id: 'msg-1',
      events: [
        {
          timestamp: '2026-04-11T12:00:00+00:00',
          phase: 'dispatch.begin',
          detail: { carrier: 'VINAPHONE' },
        },
      ],
    };

    it('successful 200', async () => {
      const mockFetch = vi.mocked(globalThis.fetch);
      mockFetch.mockResolvedValue({
        ok: true,
        status: 200,
        statusText: 'OK',
        json: () => ({ data: pipelinePayload }),
      } as Response);

      const api = createNotificationApi();
      const result = await api.getPipelineEvents('notif-id');

      expect(result).toEqual({ ok: true, value: pipelinePayload });
      expect(mockFetch).toHaveBeenCalledWith(
        'http://localhost:8001/notifications/notif-id/pipeline-events',
        expect.objectContaining({ method: 'GET' }),
      );
    });
  });
});
