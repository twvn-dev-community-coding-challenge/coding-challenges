import type {
  CreateNotificationRequest,
  DispatchRequest,
  ListNotificationsData,
  NotificationResource,
  Result,
  RetryRequest,
} from './types';

const BASE_URL = 'http://localhost:8001';

const readJson = async (response: Response): Promise<unknown> => response.json();

export const createNotificationApi = () => {
  const createNotification = async (
    request: CreateNotificationRequest,
  ): Promise<Result<NotificationResource>> => {
    try {
      const response = await fetch(`${BASE_URL}/notifications`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          Accept: 'application/json',
        },
        body: JSON.stringify(request),
      });
      const body = (await readJson(response)) as {
        readonly data?: NotificationResource;
        readonly error?: { readonly code: string; readonly message: string };
      };
      if (!response.ok) {
        if (!body.error) {
          return {
            ok: false,
            error: { code: 'HTTP_ERROR', message: 'Request failed' },
          };
        }
        return { ok: false, error: body.error };
      }
      if (body.data === undefined) {
        return {
          ok: false,
          error: { code: 'INVALID_RESPONSE', message: 'Missing data' },
        };
      }
      return { ok: true, value: body.data };
    } catch (error: unknown) {
      const message = error instanceof Error ? error.message : String(error);
      return {
        ok: false,
        error: { code: 'NETWORK_ERROR', message },
      };
    }
  };

  const dispatchNotification = async (
    notificationId: string,
    body: DispatchRequest,
  ): Promise<Result<NotificationResource>> => {
    try {
      const response = await fetch(
        `${BASE_URL}/notifications/${notificationId}/dispatch`,
        {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
            Accept: 'application/json',
          },
          body: JSON.stringify(body),
        },
      );
      const payload = (await readJson(response)) as {
        readonly data?: NotificationResource;
        readonly error?: { readonly code: string; readonly message: string };
      };
      if (!response.ok) {
        if (!payload.error) {
          return {
            ok: false,
            error: { code: 'HTTP_ERROR', message: 'Request failed' },
          };
        }
        return { ok: false, error: payload.error };
      }
      if (payload.data === undefined) {
        return {
          ok: false,
          error: { code: 'INVALID_RESPONSE', message: 'Missing data' },
        };
      }
      return { ok: true, value: payload.data };
    } catch (error: unknown) {
      const message = error instanceof Error ? error.message : String(error);
      return {
        ok: false,
        error: { code: 'NETWORK_ERROR', message },
      };
    }
  };

  const getNotification = async (
    notificationId: string,
  ): Promise<Result<NotificationResource>> => {
    try {
      const response = await fetch(
        `${BASE_URL}/notifications/${notificationId}`,
        {
          method: 'GET',
          headers: {
            Accept: 'application/json',
          },
        },
      );
      const payload = (await readJson(response)) as {
        readonly data?: NotificationResource;
        readonly error?: { readonly code: string; readonly message: string };
      };
      if (!response.ok) {
        if (!payload.error) {
          return {
            ok: false,
            error: { code: 'HTTP_ERROR', message: 'Request failed' },
          };
        }
        return { ok: false, error: payload.error };
      }
      if (payload.data === undefined) {
        return {
          ok: false,
          error: { code: 'INVALID_RESPONSE', message: 'Missing data' },
        };
      }
      return { ok: true, value: payload.data };
    } catch (error: unknown) {
      const message = error instanceof Error ? error.message : String(error);
      return {
        ok: false,
        error: { code: 'NETWORK_ERROR', message },
      };
    }
  };

  const listNotifications = async (): Promise<
    Result<readonly NotificationResource[]>
  > => {
    try {
      const response = await fetch(`${BASE_URL}/notifications`, {
        method: 'GET',
        headers: {
          Accept: 'application/json',
        },
      });
      const payload = (await readJson(response)) as {
        readonly data?: ListNotificationsData;
        readonly error?: { readonly code: string; readonly message: string };
      };
      if (!response.ok) {
        if (!payload.error) {
          return {
            ok: false,
            error: { code: 'HTTP_ERROR', message: 'Request failed' },
          };
        }
        return { ok: false, error: payload.error };
      }
      const notifications = payload.data?.notifications;
      if (!Array.isArray(notifications)) {
        return {
          ok: false,
          error: { code: 'INVALID_RESPONSE', message: 'Missing notifications' },
        };
      }
      return { ok: true, value: notifications };
    } catch (error: unknown) {
      const message = error instanceof Error ? error.message : String(error);
      return {
        ok: false,
        error: { code: 'NETWORK_ERROR', message },
      };
    }
  };

  const retryNotification = async (
    notificationId: string,
    body: RetryRequest,
  ): Promise<Result<NotificationResource>> => {
    try {
      const response = await fetch(
        `${BASE_URL}/notifications/${notificationId}/retry`,
        {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
            Accept: 'application/json',
          },
          body: JSON.stringify(body),
        },
      );
      const payload = (await readJson(response)) as {
        readonly data?: NotificationResource;
        readonly error?: { readonly code: string; readonly message: string };
      };
      if (!response.ok) {
        if (!payload.error) {
          return {
            ok: false,
            error: { code: 'HTTP_ERROR', message: 'Request failed' },
          };
        }
        return { ok: false, error: payload.error };
      }
      if (payload.data === undefined) {
        return {
          ok: false,
          error: { code: 'INVALID_RESPONSE', message: 'Missing data' },
        };
      }
      return { ok: true, value: payload.data };
    } catch (error: unknown) {
      const message = error instanceof Error ? error.message : String(error);
      return {
        ok: false,
        error: { code: 'NETWORK_ERROR', message },
      };
    }
  };

  return {
    createNotification,
    dispatchNotification,
    getNotification,
    listNotifications,
    retryNotification,
  };
};
