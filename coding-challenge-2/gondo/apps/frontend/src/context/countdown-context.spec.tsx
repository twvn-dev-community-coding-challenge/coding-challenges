import { act, renderHook } from '@testing-library/react';
import { describe, it, expect, vi, afterEach } from 'vitest';
import type { ReactNode } from 'react';

import {
  CountdownProvider,
  DEFAULT_COUNTDOWN_DURATION,
  useCountdown,
} from './countdown-context';

const wrapper = ({ children }: { readonly children: ReactNode }) => (
  <CountdownProvider>{children}</CountdownProvider>
);

describe('CountdownProvider / useCountdown', () => {
  afterEach(() => {
    vi.restoreAllMocks();
  });

  it('throws when useCountdown is used outside CountdownProvider', () => {
    expect(() => {
      renderHook(() => useCountdown());
    }).toThrow(/CountdownProvider/i);
  });

  it('startCountdown adds an entry', () => {
    const { result } = renderHook(() => useCountdown(), { wrapper });

    act(() => {
      result.current.startCountdown({
        notificationId: 'n1',
        otp: '111111',
        startedAt: 1_000,
        durationSeconds: DEFAULT_COUNTDOWN_DURATION,
      });
    });

    expect(result.current.entries.has('n1')).toBe(true);
    expect(result.current.entries.get('n1')?.otp).toBe('111111');
  });

  it('getRemainingSeconds returns full duration when time has not advanced', () => {
    vi.spyOn(Date, 'now').mockReturnValue(10_000);

    const { result } = renderHook(() => useCountdown(), { wrapper });

    act(() => {
      result.current.startCountdown({
        notificationId: 'n1',
        otp: '111111',
        startedAt: 10_000,
        durationSeconds: 100,
      });
    });

    expect(result.current.getRemainingSeconds('n1')).toBe(100);
  });

  it('getRemainingSeconds decreases as wall clock advances', () => {
    const { result } = renderHook(() => useCountdown(), { wrapper });

    act(() => {
      result.current.startCountdown({
        notificationId: 'n1',
        otp: '111111',
        startedAt: 0,
        durationSeconds: 100,
      });
    });

    vi.spyOn(Date, 'now').mockReturnValue(45_000);

    expect(result.current.getRemainingSeconds('n1')).toBe(55);
  });

  it('getRemainingSeconds returns 0 when elapsed exceeds duration', () => {
    const { result } = renderHook(() => useCountdown(), { wrapper });

    act(() => {
      result.current.startCountdown({
        notificationId: 'n1',
        otp: '111111',
        startedAt: 0,
        durationSeconds: 60,
      });
    });

    vi.spyOn(Date, 'now').mockReturnValue(120_000);

    expect(result.current.getRemainingSeconds('n1')).toBe(0);
    expect(result.current.isExpired('n1')).toBe(true);
  });

  it('isExpired is false while time remains', () => {
    vi.spyOn(Date, 'now').mockReturnValue(0);

    const { result } = renderHook(() => useCountdown(), { wrapper });

    act(() => {
      result.current.startCountdown({
        notificationId: 'n1',
        otp: '111111',
        startedAt: 0,
        durationSeconds: 60,
      });
    });

    vi.spyOn(Date, 'now').mockReturnValue(30_000);

    expect(result.current.isExpired('n1')).toBe(false);
  });

  it('resetCountdown updates otp and restarts startedAt', () => {
    vi.spyOn(Date, 'now').mockReturnValue(100_000);

    const { result } = renderHook(() => useCountdown(), { wrapper });

    act(() => {
      result.current.startCountdown({
        notificationId: 'n1',
        otp: '111111',
        startedAt: 0,
        durationSeconds: 60,
      });
    });

    act(() => {
      result.current.resetCountdown('n1', '222222');
    });

    expect(result.current.entries.get('n1')?.otp).toBe('222222');
    expect(result.current.entries.get('n1')?.startedAt).toBe(100_000);
    expect(result.current.getRemainingSeconds('n1')).toBe(60);
  });

  it('removeCountdown removes an entry', () => {
    const { result } = renderHook(() => useCountdown(), { wrapper });

    act(() => {
      result.current.startCountdown({
        notificationId: 'n1',
        otp: '111111',
        startedAt: 0,
        durationSeconds: 60,
      });
    });

    act(() => {
      result.current.removeCountdown('n1');
    });

    expect(result.current.entries.has('n1')).toBe(false);
  });
});
