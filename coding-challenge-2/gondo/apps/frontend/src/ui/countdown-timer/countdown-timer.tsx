import { keyframes } from '@emotion/react';
import { Typography } from '@mui/material';
import { useEffect, useRef, useState } from 'react';

interface CountdownTimerProps {
  readonly durationSeconds: number;
  readonly onExpire: () => void;
}

const formatMmSs = (totalSeconds: number): string => {
  const safe = Math.max(0, totalSeconds);
  const minutes = Math.floor(safe / 60);
  const seconds = safe % 60;
  return `${String(minutes).padStart(2, '0')}:${String(seconds).padStart(2, '0')}`;
};

const pulse = keyframes`
  0%,
  100% {
    transform: scale(1);
    opacity: 1;
  }

  50% {
    transform: scale(1.03);
    opacity: 0.92;
  }
`;

export const CountdownTimer = ({
  durationSeconds,
  onExpire,
}: CountdownTimerProps) => {
  const [remaining, setRemaining] = useState(durationSeconds);
  const expiredRef = useRef(false);
  const onExpireRef = useRef(onExpire);

  onExpireRef.current = onExpire;

  useEffect(() => {
    const id = window.setInterval(() => {
      setRemaining((previous) => {
        if (previous <= 0) {
          return 0;
        }
        if (previous === 1) {
          if (!expiredRef.current) {
            expiredRef.current = true;
            onExpireRef.current();
          }
          return 0;
        }
        return previous - 1;
      });
    }, 1000);

    return () => {
      window.clearInterval(id);
    };
  }, []);

  const isLowTime = remaining > 0 && remaining < 30;

  return (
    <Typography
      component="span"
      role="timer"
      aria-live="polite"
      aria-label={`Time remaining: ${formatMmSs(remaining)}`}
      sx={{
        display: 'inline-block',
        fontFamily:
          'ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, "Liberation Mono", "Courier New", monospace',
        fontSize: '2rem',
        fontWeight: 600,
        letterSpacing: '0.06em',
        color: isLowTime ? 'error.main' : 'text.primary',
        transition: 'color 0.25s ease, transform 0.25s ease',
        ...(isLowTime && {
          animation: `${pulse} 1.1s ease-in-out infinite`,
        }),
      }}
    >
      {formatMmSs(remaining)}
    </Typography>
  );
};
