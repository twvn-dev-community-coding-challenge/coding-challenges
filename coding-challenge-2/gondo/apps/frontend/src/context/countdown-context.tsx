import {
  createContext,
  useCallback,
  useContext,
  useMemo,
  useState,
  type ReactNode,
} from 'react';

export const DEFAULT_COUNTDOWN_DURATION = 120;

export interface CountdownEntry {
  readonly notificationId: string;
  readonly otp: string;
  readonly startedAt: number;
  readonly durationSeconds: number;
}

export interface CountdownContextValue {
  readonly entries: ReadonlyMap<string, CountdownEntry>;
  readonly startCountdown: (entry: CountdownEntry) => void;
  readonly resetCountdown: (notificationId: string, newOtp: string) => void;
  readonly removeCountdown: (notificationId: string) => void;
  readonly getRemainingSeconds: (notificationId: string) => number;
  readonly isExpired: (notificationId: string) => boolean;
}

const CountdownContext = createContext<CountdownContextValue | null>(null);

export const useCountdown = (): CountdownContextValue => {
  const value = useContext(CountdownContext);
  if (value === null) {
    throw new Error('useCountdown must be used within a CountdownProvider');
  }
  return value;
};

interface CountdownProviderProps {
  readonly children: ReactNode;
}

export const CountdownProvider = ({ children }: CountdownProviderProps) => {
  const [entries, setEntries] = useState<Map<string, CountdownEntry>>(
    () => new Map(),
  );

  const startCountdown = useCallback((entry: CountdownEntry) => {
    setEntries((previous) => {
      const next = new Map(previous);
      next.set(entry.notificationId, entry);
      return next;
    });
  }, []);

  const resetCountdown = useCallback(
    (notificationId: string, newOtp: string) => {
      setEntries((previous) => {
        const existing = previous.get(notificationId);
        if (existing === undefined) {
          return previous;
        }
        const next = new Map(previous);
        next.set(notificationId, {
          ...existing,
          otp: newOtp,
          startedAt: Date.now(),
        });
        return next;
      });
    },
    [],
  );

  const removeCountdown = useCallback((notificationId: string) => {
    setEntries((previous) => {
      if (!previous.has(notificationId)) {
        return previous;
      }
      const next = new Map(previous);
      next.delete(notificationId);
      return next;
    });
  }, []);

  const getRemainingSeconds = useCallback(
    (notificationId: string): number => {
      const entry = entries.get(notificationId);
      if (entry === undefined) {
        return 0;
      }
      const elapsed = Math.floor((Date.now() - entry.startedAt) / 1000);
      return Math.max(0, entry.durationSeconds - elapsed);
    },
    [entries],
  );

  const isExpired = useCallback(
    (notificationId: string): boolean =>
      getRemainingSeconds(notificationId) === 0,
    [getRemainingSeconds],
  );

  const value = useMemo(
    (): CountdownContextValue => ({
      entries,
      startCountdown,
      resetCountdown,
      removeCountdown,
      getRemainingSeconds,
      isExpired,
    }),
    [
      entries,
      startCountdown,
      resetCountdown,
      removeCountdown,
      getRemainingSeconds,
      isExpired,
    ],
  );

  return (
    <CountdownContext.Provider value={value}>{children}</CountdownContext.Provider>
  );
};
