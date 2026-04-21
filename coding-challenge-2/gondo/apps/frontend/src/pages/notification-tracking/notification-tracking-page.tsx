import RefreshIcon from '@mui/icons-material/Refresh';
import SearchIcon from '@mui/icons-material/Search';
import {
  Alert,
  Box,
  Button,
  Card,
  CardContent,
  Chip,
  CircularProgress,
  Dialog,
  DialogActions,
  DialogContent,
  DialogTitle,
  InputAdornment,
  Link,
  Paper,
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TableRow,
  TextField,
  Typography,
} from '@mui/material';
import { useCallback, useEffect, useMemo, useState } from 'react';

import { createNotificationApi, generateSixDigitOtp } from '@gondo/ts-core';
import type {
  NotificationResource,
  PipelineEvent,
  PipelineEventsData,
} from '@gondo/ts-core';

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
  n.selected_provider_code ?? n.selected_provider_id ?? '—';

const sortByCreatedAtDesc = (
  items: readonly NotificationResource[],
): NotificationResource[] =>
  [...items].sort(
    (a, b) =>
      new Date(b.created_at).getTime() - new Date(a.created_at).getTime(),
  );

interface NotificationRowActionsProps {
  readonly notificationId: string;
  readonly state: string;
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
  state,
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
        aria-label={`Retry notification (state: ${state})`}
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

interface PipelineLogsLinkProps {
  readonly notificationId: string;
  readonly api: ReturnType<typeof createNotificationApi>;
}

const PipelineLogsLink = ({ notificationId, api }: PipelineLogsLinkProps) => {
  const [open, setOpen] = useState(false);
  const [loading, setLoading] = useState(false);
  const [errorMessage, setErrorMessage] = useState<string | null>(null);
  const [payload, setPayload] = useState<PipelineEventsData | null>(null);
  const [searchQuery, setSearchQuery] = useState('');
  const [selectedCategories, setSelectedCategories] = useState<
    ReadonlySet<string>
  >(new Set());

  const getPhaseCategory = useCallback((phase: string): string => {
    const dot = phase.indexOf('.');
    return dot === -1 ? phase : phase.slice(0, dot);
  }, []);

  const getPhaseActionLabel = useCallback((phase: string): string => {
    const dot = phase.indexOf('.');
    if (dot === -1) {
      return phase;
    }
    return phase.slice(dot + 1);
  }, []);

  const formatDetailValue = useCallback((value: unknown): string => {
    if (value === null || value === undefined) {
      return String(value);
    }
    if (typeof value === 'object') {
      return JSON.stringify(value);
    }
    return String(value);
  }, []);

  const load = useCallback(async () => {
    setLoading(true);
    setErrorMessage(null);
    const result = await api.getPipelineEvents(notificationId);
    setLoading(false);
    if (!result.ok) {
      setErrorMessage(result.error.message);
      setPayload(null);
      setSelectedCategories(new Set());
      return;
    }
    const value = result.value;
    setPayload(value);
    if (value.events.length > 0) {
      setSelectedCategories(
        new Set(value.events.map((e) => getPhaseCategory(e.phase))),
      );
    } else {
      setSelectedCategories(new Set());
    }
  }, [api, getPhaseCategory, notificationId]);

  useEffect(() => {
    if (open) {
      void load();
    }
  }, [open, load]);

  const uniqueCategories = useMemo(() => {
    if (payload === null) {
      return [] as string[];
    }
    const seen = new Set<string>();
    const ordered: string[] = [];
    for (const ev of payload.events) {
      const c = getPhaseCategory(ev.phase);
      if (!seen.has(c)) {
        seen.add(c);
        ordered.push(c);
      }
    }
    return ordered;
  }, [getPhaseCategory, payload]);

  const filteredEvents = useMemo(() => {
    if (payload === null) {
      return [] as readonly PipelineEvent[];
    }
    const q = searchQuery.trim().toLowerCase();
    return payload.events.filter((ev) => {
      if (!selectedCategories.has(getPhaseCategory(ev.phase))) {
        return false;
      }
      if (q.length === 0) {
        return true;
      }
      if (ev.phase.toLowerCase().includes(q)) {
        return true;
      }
      if (ev.timestamp.toLowerCase().includes(q)) {
        return true;
      }
      for (const [key, value] of Object.entries(ev.detail)) {
        if (key.toLowerCase().includes(q)) {
          return true;
        }
        if (formatDetailValue(value).toLowerCase().includes(q)) {
          return true;
        }
      }
      return false;
    });
  }, [formatDetailValue, getPhaseCategory, payload, searchQuery, selectedCategories]);

  const handleClose = () => {
    setOpen(false);
    setPayload(null);
    setErrorMessage(null);
    setSearchQuery('');
    setSelectedCategories(new Set());
  };

  const toggleCategory = (category: string): void => {
    setSelectedCategories((prev) => {
      const next = new Set(prev);
      if (next.has(category)) {
        next.delete(category);
      } else {
        next.add(category);
      }
      return next;
    });
  };

  return (
    <>
      <Link
        component="button"
        type="button"
        variant="body2"
        onClick={() => setOpen(true)}
        sx={{ cursor: 'pointer', whiteSpace: 'nowrap' }}
      >
        Pipeline logs
      </Link>
      <Dialog
        open={open}
        onClose={handleClose}
        maxWidth="md"
        fullWidth
        aria-labelledby="pipeline-logs-dialog-title"
      >
        <DialogTitle id="pipeline-logs-dialog-title">
          SMS pipeline (runtime)
        </DialogTitle>
        <DialogContent dividers>
          <Typography variant="body2" color="text.secondary" sx={{ mb: 1 }}>
            Notification <code>{notificationId}</code>
            {payload?.message_id ? (
              <>
                {' '}
                · message_id <code>{payload.message_id}</code>
              </>
            ) : null}
          </Typography>
          {loading ? (
            <Box sx={{ display: 'flex', justifyContent: 'center', py: 3 }}>
              <CircularProgress size={28} aria-busy="true" />
            </Box>
          ) : null}
          {errorMessage !== null ? (
            <Alert severity="error" sx={{ mb: 1 }}>
              {errorMessage}
            </Alert>
          ) : null}
          {!loading && payload !== null && errorMessage === null ? (
            payload.events.length === 0 ? (
              <Typography variant="body2" color="text.secondary">
                No pipeline events recorded yet for this notification in this
                process.
              </Typography>
            ) : (
              <Box>
                <TextField
                  fullWidth
                  size="small"
                  placeholder="Search by phase or detail…"
                  value={searchQuery}
                  onChange={(e) => setSearchQuery(e.target.value)}
                  sx={{ mb: 1.5 }}
                  slotProps={{
                    input: {
                      startAdornment: (
                        <InputAdornment position="start">
                          <SearchIcon fontSize="small" color="action" />
                        </InputAdornment>
                      ),
                    },
                  }}
                />
                <Box
                  sx={{
                    display: 'flex',
                    flexWrap: 'wrap',
                    gap: 0.75,
                    alignItems: 'center',
                    mb: 1,
                  }}
                >
                  <Typography
                    variant="caption"
                    color="text.secondary"
                    sx={{ mr: 0.5 }}
                  >
                    Category:
                  </Typography>
                  {uniqueCategories.map((cat) => {
                    const selected = selectedCategories.has(cat);
                    return (
                      <Chip
                        key={cat}
                        label={cat}
                        size="small"
                        variant={selected ? 'filled' : 'outlined'}
                        color={selected ? 'primary' : 'default'}
                        onClick={() => toggleCategory(cat)}
                      />
                    );
                  })}
                </Box>
                <Typography variant="caption" color="text.secondary" sx={{ mb: 1, display: 'block' }}>
                  Showing {filteredEvents.length} of {payload.events.length}{' '}
                  events
                </Typography>
                <Box
                  sx={{
                    maxHeight: 500,
                    overflow: 'auto',
                    pr: 0.5,
                  }}
                >
                  {filteredEvents.length === 0 ? (
                    <Typography variant="body2" color="text.secondary">
                      No events match the current filters.
                    </Typography>
                  ) : (
                    <Box sx={{ display: 'flex', flexDirection: 'column', gap: 1.5 }}>
                      {filteredEvents.map((ev, index) => {
                        const category = getPhaseCategory(ev.phase);
                        const actionLabel = getPhaseActionLabel(ev.phase);
                        const detailKeys = Object.keys(ev.detail);
                        const borderLeftColor = (() => {
                          switch (category) {
                            case 'state':
                              return 'primary.main';
                            case 'dispatch':
                              return 'warning.main';
                            case 'bus':
                              return 'secondary.main';
                            case 'retry':
                              return 'error.main';
                            default:
                              return 'divider';
                          }
                        })();
                        return (
                          <Card
                            key={`${ev.timestamp}-${ev.phase}-${String(index)}`}
                            variant="outlined"
                            sx={{
                              borderLeftWidth: 4,
                              borderLeftStyle: 'solid',
                              borderLeftColor,
                            }}
                          >
                            <CardContent
                              sx={{
                                '&:last-child': { pb: 2 },
                                py: 1.5,
                                px: 2,
                              }}
                            >
                              <Box
                                sx={{
                                  display: 'flex',
                                  flexWrap: 'wrap',
                                  alignItems: 'center',
                                  gap: 1,
                                  mb: 1,
                                }}
                              >
                                <Chip
                                  label={category}
                                  size="small"
                                  sx={{ fontWeight: 600 }}
                                />
                                <Typography
                                  variant="subtitle2"
                                  component="span"
                                  sx={{ flex: '1 1 auto', minWidth: 0 }}
                                >
                                  {actionLabel}
                                </Typography>
                                <Typography
                                  variant="caption"
                                  color="text.secondary"
                                  sx={{ whiteSpace: 'nowrap' }}
                                >
                                  {formatDateTime(ev.timestamp)}
                                </Typography>
                              </Box>
                              {detailKeys.length > 0 ? (
                                <Box
                                  component="dl"
                                  sx={{
                                    m: 0,
                                    display: 'grid',
                                    gridTemplateColumns: {
                                      xs: '1fr',
                                      sm: 'minmax(100px, 140px) 1fr',
                                    },
                                    gap: { xs: 0.5, sm: 1 },
                                    rowGap: 0.75,
                                  }}
                                >
                                  {detailKeys.map((key) => (
                                    <Box
                                      key={key}
                                      sx={{
                                        display: 'contents',
                                      }}
                                    >
                                      <Typography
                                        component="dt"
                                        variant="caption"
                                        color="text.secondary"
                                        sx={{ fontWeight: 600 }}
                                      >
                                        {key}
                                      </Typography>
                                      <Typography
                                        component="dd"
                                        variant="body2"
                                        sx={{ m: 0, wordBreak: 'break-word' }}
                                      >
                                        {formatDetailValue(ev.detail[key])}
                                      </Typography>
                                    </Box>
                                  ))}
                                </Box>
                              ) : null}
                            </CardContent>
                          </Card>
                        );
                      })}
                    </Box>
                  )}
                </Box>
              </Box>
            )
          ) : null}
        </DialogContent>
        <DialogActions>
          <Button
            type="button"
            onClick={() => void load()}
            disabled={loading}
            startIcon={<RefreshIcon />}
          >
            Refresh
          </Button>
          <Button type="button" onClick={handleClose}>
            Close
          </Button>
        </DialogActions>
      </Dialog>
    </>
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
            sx={{ minWidth: 980 }}
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
                <TableCell sx={{ whiteSpace: 'nowrap' }} align="center">
                  Logs
                </TableCell>
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
                  <TableCell align="center" sx={{ verticalAlign: 'middle' }}>
                    <PipelineLogsLink
                      notificationId={n.notification_id}
                      api={api}
                    />
                  </TableCell>
                  <TableCell
                    sx={{ minWidth: 180, verticalAlign: 'middle' }}
                    align="center"
                  >
                    <NotificationRowActions
                      notificationId={n.notification_id}
                      state={n.state}
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
