import { screen, act } from '@testing-library/react';

import { renderWithTheme } from '../../test-utils';
import { CountdownTimer } from './countdown-timer';

describe('CountdownTimer', () => {
  it('renders initial duration in mm:ss format', () => {
    renderWithTheme(<CountdownTimer durationSeconds={120} onExpire={vi.fn()} />);
    expect(screen.getByText('02:00')).toBeTruthy();
  });

  it('decrements every second', () => {
    vi.useFakeTimers();
    renderWithTheme(<CountdownTimer durationSeconds={5} onExpire={vi.fn()} />);
    expect(screen.getByText('00:05')).toBeTruthy();
    act(() => {
      vi.advanceTimersByTime(1000);
    });
    expect(screen.getByText('00:04')).toBeTruthy();
    act(() => {
      vi.advanceTimersByTime(1000);
    });
    expect(screen.getByText('00:03')).toBeTruthy();
    vi.useRealTimers();
  });

  it('calls onExpire when reaching zero', () => {
    vi.useFakeTimers();
    const onExpire = vi.fn();
    renderWithTheme(<CountdownTimer durationSeconds={2} onExpire={onExpire} />);
    act(() => {
      vi.advanceTimersByTime(2000);
    });
    expect(onExpire).toHaveBeenCalledTimes(1);
    vi.useRealTimers();
  });

  it('does not call onExpire more than once', () => {
    vi.useFakeTimers();
    const onExpire = vi.fn();
    renderWithTheme(<CountdownTimer durationSeconds={1} onExpire={onExpire} />);
    act(() => {
      vi.advanceTimersByTime(3000);
    });
    expect(onExpire).toHaveBeenCalledTimes(1);
    vi.useRealTimers();
  });

  it('has accessible time display with role="timer"', () => {
    renderWithTheme(<CountdownTimer durationSeconds={60} onExpire={vi.fn()} />);
    expect(screen.getByRole('timer')).toBeTruthy();
  });
});
