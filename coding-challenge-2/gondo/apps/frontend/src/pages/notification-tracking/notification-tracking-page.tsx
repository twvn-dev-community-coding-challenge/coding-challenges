import RefreshIcon from '@mui/icons-material/Refresh';
import {
  Alert,
  Box,
  Button,
  CircularProgress,
  Paper,
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TableRow,
  Typography,
} from '@mui/material';
import { useCallback, useEffect, useMemo, useState } from 'react';

import { createNotificationApi, generateSixDigitOtp } from '@gondo/ts-core';
import type { NotificationResource } from '@gondo/ts-core';

import { DEFAULT_COUNTDOWN_DURATION, useCountdown } from '../../context';
import { CountdownTimer } from '../../ui/countdown-timer/countdown-timer';
import { NotificationStatusBadge } from '../../ui/notification-status-badge/notification-status-badge';
import { OtpDisplay } from '../../ui/otp-display/otp-display';

const formatDateTime = (iso: string): string => {
  const d = new Date(iso);
  if (Number.isNaN(d.getTime())) {
    return iso;
  }
  return d.toLocaleString(undefined, {
    dateStyle: 'short',
    timeStyle: 'medium',
  });
};

const getPhone = (n: NotificationResource): string =>
  n.channel_payload.phone_number ?? '—';

const getProvider = (n: NotificationResource): string =>
  n.selected_provider_id ?? '—';

const sortByCreatedAtDesc = (
  items: readonly NotificationResource[],
): NotificationResource[] =>
  [...items].sort(
    (a, b) =>
      new Date(b.created_at).getTime() - new Date(a.created_at).getTime(),
  );

interface NotificationRowActionsProps {
  readonly notificationId: string;
  readonly createdAt: string;
  readonly durationSeconds?: number;
  readonly api: ReturnType<typeof createNotificationApi>;
  readonly onRetryComplete: () => void | Promise<void>;
}

const computeStartMs = (
  createdAt: string,
  contextStartedAt: number | undefined,
): number => {
  const createdAtMs = new Date(createdAt).getTime();
  if (contextStartedAt !== undefined && contextStartedAt > createdAtMs) {
    return contextStartedAt;
  }
  return createdAtMs;
};

const NotificationRowActions = ({
  notificationId,
  createdAt,
  durationSeconds = DEFAULT_COUNTDOWN_DURATION,
  api,
  onRetryComplete,
}: NotificationRowActionsProps) => {
  const { entries, resetCountdown, startCountdown } = useCountdown();
  const [retryError, setRetryError] = useState<string | null>(null);
  const [retrying, setRetrying] = useState(false);
  const [timerExpired, setTimerExpired] = useState(false);

  const entry = entries.get(notificationId);

  const startMs = useMemo(
    () => computeStartMs(createdAt, entry?.startedAt),
    [createdAt, entry?.startedAt],
  );

  const computedRemaining = Math.max(
    0,
    durationSeconds - Math.floor((Date.now() - startMs) / 1000),
  );

  useEffect(() => {
    setTimerExpired(false);
  }, [entry?.otp, entry?.startedAt, notificationId, createdAt]);

  const handleRetry = useCallback(async () => {
    setRetrying(true);
    setRetryError(null);
    const newOtp = generateSixDigitOtp();
    const result = await api.retryNotification(notificationId, {
      as_of: new Date().toISOString(),
    });
    if (!result.ok) {
      setRetryError(result.error.message);
      setRetrying(false);
      return;
    }
    if (entry !== undefined) {
      resetCountdown(notificationId, newOtp);
    } else {
      startCountdown({
        notificationId,
        otp: newOtp,
        startedAt: Date.now(),
        durationSeconds,
      });
    }
    setRetrying(false);
    await onRetryComplete();
  }, [
    api,
    notificationId,
    onRetryComplete,
    resetCountdown,
    startCountdown,
    entry,
    durationSeconds,
  ]);

  const handleTimerExpire = () => {
    queueMicrotask(() => {
      setTimerExpired(true);
    });
  };

  const expired = computedRemaining === 0 || timerExpired;

  if (!expired && computedRemaining > 0) {
    return (
      <Box
        sx={{
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          flexWrap: 'wrap',
          gap: 1,
        }}
      >
        <Box
          sx={{
            '& [role="timer"]': {
              fontSize: '1.25rem',
            },
          }}
        >
          <CountdownTimer
            key={`${notificationId}-${startMs}-${entry?.otp ?? ''}`}
            durationSeconds={computedRemaining}
            onExpire={handleTimerExpire}
          />
        </Box>
        {entry !== undefined && (
          <Box
            sx={{
              transform: 'scale(0.5)',
              transformOrigin: 'left center',
            }}
          >
            <OtpDisplay otp={entry.otp} />
          </Box>
        )}
      </Box>
    );
  }

  return (
    <Box
      sx={{
        display: 'flex',
        flexDirection: 'column',
        alignItems: 'center',
        gap: 1,
      }}
    >
      <Button
        type="button"
        variant="outlined"
        size="small"
        onClick={() => void handleRetry()}
        disabled={retrying}
      >
        Retry
      </Button>
      {retryError !== null ? (
        <Alert
          severity="error"
          sx={{ py: 0, px: 1, fontSize: '0.75rem', maxWidth: 220 }}
        >
          {retryError}
        </Alert>
      ) : null}
    </Box>
  );
};

export const NotificationTrackingPage = () => {
  const api = useMemo(() => createNotificationApi(), []);
  const [rows, setRows] = useState<readonly NotificationResource[]>([]);
  const [loading, setLoading] = useState(true);
  const [errorMessage, setErrorMessage] = useState<string | null>(null);

  const load = useCallback(async () => {
    setLoading(true);
    setErrorMessage(null);
    const result = await api.listNotifications();
    setLoading(false);
    if (!result.ok) {
      setErrorMessage(result.error.message);
      setRows([]);
      return;
    }
    setRows(sortByCreatedAtDesc(result.value));
  }, [api]);

  useEffect(() => {
    void load();
  }, [load]);

  return (
    <Box component="main">
      <Box
        sx={{
          display: 'flex',
          flexWrap: 'wrap',
          alignItems: 'center',
          justifyContent: 'space-between',
          gap: 2,
          mb: 2,
        }}
      >
        <Typography variant="h4" component="h1">
          Notification Tracking
        </Typography>
        <Button
          type="button"
          variant="outlined"
          startIcon={<RefreshIcon />}
          onClick={() => void load()}
          disabled={loading}
        >
          Refresh
        </Button>
      </Box>

      {errorMessage !== null ? (
        <Alert severity="error" sx={{ mb: 2 }}>
          Failed to load notifications: {errorMessage}
        </Alert>
      ) : null}

      {loading ? (
        <Box
          sx={{
            display: 'flex',
            justifyContent: 'center',
            py: 6,
          }}
          aria-busy="true"
        >
          <CircularProgress />
        </Box>
      ) : null}

      {!loading && errorMessage === null ? (
        <TableContainer
          component={Paper}
          elevation={0}
          variant="outlined"
          sx={{ width: '100%', overflowX: 'auto' }}
        >
          <Table
            size="small"
            aria-label="Notification tracking"
            sx={{ minWidth: 900 }}
          >
            <TableHead>
              <TableRow>
                <TableCell sx={{ whiteSpace: 'nowrap' }}>Status</TableCell>
                <TableCell sx={{ whiteSpace: 'nowrap' }}>Message ID</TableCell>
                <TableCell sx={{ whiteSpace: 'nowrap' }}>Recipient</TableCell>
                <TableCell sx={{ whiteSpace: 'nowrap' }}>Phone</TableCell>
                <TableCell sx={{ whiteSpace: 'nowrap' }}>Provider</TableCell>
                <TableCell sx={{ whiteSpace: 'nowrap' }} align="right">
                  Attempt
                </TableCell>
                <TableCell sx={{ whiteSpace: 'nowrap' }}>Created At</TableCell>
                <TableCell sx={{ whiteSpace: 'nowrap' }}>Updated At</TableCell>
                <TableCell
                  sx={{ whiteSpace: 'nowrap', minWidth: 180 }}
                  align="center"
                >
                  Actions
                </TableCell>
              </TableRow>
            </TableHead>
            <TableBody>
              {rows.map((n) => (
                <TableRow key={n.notification_id}>
                  <TableCell sx={{ whiteSpace: 'nowrap' }}>
                    <NotificationStatusBadge statusLabel={n.state} />
                  </TableCell>
                  <TableCell
                    sx={{
                      maxWidth: 200,
                      overflow: 'hidden',
                      textOverflow: 'ellipsis',
                      whiteSpace: 'nowrap',
                    }}
                    title={n.message_id}
                  >
                    {n.message_id}
                  </TableCell>
                  <TableCell
                    sx={{
                      maxWidth: 160,
                      overflow: 'hidden',
                      textOverflow: 'ellipsis',
                      whiteSpace: 'nowrap',
                    }}
                    title={n.recipient}
                  >
                    {n.recipient}
                  </TableCell>
                  <TableCell sx={{ whiteSpace: 'nowrap' }}>
                    {getPhone(n)}
                  </TableCell>
                  <TableCell sx={{ whiteSpace: 'nowrap' }}>
                    {getProvider(n)}
                  </TableCell>
                  <TableCell align="right">{n.attempt}</TableCell>
                  <TableCell sx={{ whiteSpace: 'nowrap' }}>
                    {formatDateTime(n.created_at)}
                  </TableCell>
                  <TableCell sx={{ whiteSpace: 'nowrap' }}>
                    {formatDateTime(n.updated_at)}
                  </TableCell>
                  <TableCell
                    sx={{ minWidth: 180, verticalAlign: 'middle' }}
                    align="center"
                  >
                    <NotificationRowActions
                      notificationId={n.notification_id}
                      createdAt={n.created_at}
                      api={api}
                      onRetryComplete={load}
                    />
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </TableContainer>
      ) : null}
    </Box>
  );
};
