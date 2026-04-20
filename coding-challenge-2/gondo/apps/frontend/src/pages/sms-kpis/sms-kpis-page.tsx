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
import { useCallback, useEffect, useState } from 'react';

import { createNotificationApi } from '@gondo/ts-core';
import type { SmsKpisData } from '@gondo/ts-core';

const formatPct = (v: number | null | undefined): string => {
  if (v === null || v === undefined) {
    return '—';
  }
  return `${(v * 100).toFixed(1)}%`;
};

const formatCost = (n: number): string => n.toFixed(4);

export const SmsKpisPage = () => {
  const [data, setData] = useState<SmsKpisData | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);

  const load = useCallback(async () => {
    setLoading(true);
    setError(null);
    const api = createNotificationApi();
    const res = await api.getSmsKpis();
    if (!res.ok) {
      setError(res.error.message);
      setData(null);
    } else {
      setData(res.value);
    }
    setLoading(false);
  }, []);

  useEffect(() => {
    void load();
  }, [load]);

  if (loading) {
    return (
      <Box
        data-testid="kpis-loading"
        sx={{
          display: 'flex',
          justifyContent: 'center',
          alignItems: 'center',
          minHeight: 200,
        }}
      >
        <CircularProgress />
      </Box>
    );
  }

  if (error !== null) {
    return (
      <Box>
        <Typography variant="h4" component="h1" gutterBottom>
          SMS observability (KPIs)
        </Typography>
        <Alert severity="error">Failed to load KPIs: {error}</Alert>
        <Button
          startIcon={<RefreshIcon />}
          onClick={() => void load()}
          sx={{ mt: 2 }}
        >
          Retry
        </Button>
      </Box>
    );
  }

  if (data === null) {
    return null;
  }

  const o = data.overall;

  return (
    <Box>
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
          SMS observability (KPIs)
        </Typography>
        <Button
          startIcon={<RefreshIcon />}
          onClick={() => void load()}
          variant="outlined"
        >
          Refresh
        </Button>
      </Box>

      <Alert severity="info" sx={{ mb: 2 }}>
        {data.source === 'in_memory_notifications'
          ? 'Data is aggregated from the running notification-service in-memory store (resets on process restart or tests). '
          : ''}
        {data.currency_note}{' '}
        {((data.created_at_filter?.from ?? null) !== null ||
          (data.created_at_filter?.to ?? null) !== null) && (
          <>
            <strong>created_at window:</strong> from{' '}
            {data.created_at_filter?.from ?? '—'} to {data.created_at_filter?.to ?? '—'}.
          </>
        )}
      </Alert>

      <Paper sx={{ p: 2, mb: 3 }} elevation={1}>
        <Typography variant="h6" gutterBottom>
          Overall
        </Typography>
        <Typography variant="body2" color="text.secondary">
          Total notifications: <strong>{o.total_notifications}</strong>
          {' · '}In flight: <strong>{o.in_flight}</strong>
          {' · '}Send-success: <strong>{o.send_success}</strong>
          {' · '}Send-failed: <strong>{o.send_failed}</strong>
          {' · '}Carrier-rejected: <strong>{o.carrier_rejected}</strong>
        </Typography>
        <Typography variant="body2" sx={{ mt: 1 }}>
          Terminal success rate (vs failed+rejected):{' '}
          <strong>{formatPct(o.terminal_success_rate)}</strong>
          {' · '}Estimated Σ: <strong>{formatCost(o.total_estimated_cost)}</strong>
          {' · '}Actual Σ: <strong>{formatCost(o.total_actual_cost)}</strong>
        </Typography>
      </Paper>

      <Typography variant="h6" gutterBottom>
        By provider (volume & cost)
      </Typography>
      <TableContainer component={Paper} sx={{ mb: 3 }} elevation={1}>
        <Table size="small" aria-label="KPI by provider">
          <TableHead>
            <TableRow>
              <TableCell>Provider</TableCell>
              <TableCell align="right">Volume</TableCell>
              <TableCell align="right">Σ estimated</TableCell>
              <TableCell align="right">Σ actual</TableCell>
              <TableCell align="right">Success %</TableCell>
            </TableRow>
          </TableHead>
          <TableBody>
            {data.by_provider.map((row) => (
              <TableRow key={row.provider_id}>
                <TableCell>{row.provider_id}</TableCell>
                <TableCell align="right">{row.volume}</TableCell>
                <TableCell align="right">
                  {formatCost(row.total_estimated_cost)}
                </TableCell>
                <TableCell align="right">
                  {formatCost(row.total_actual_cost)}
                </TableCell>
                <TableCell align="right">
                  {formatPct(row.terminal_success_rate)}
                </TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      </TableContainer>

      <Typography variant="h6" gutterBottom>
        By calling domain (US2 —{' '}
        <code>X-Calling-Domain</code> header on create)
      </Typography>
      <Typography variant="body2" color="text.secondary" sx={{ mb: 1 }}>
        Notifications created without the header roll into “unattributed”.
      </Typography>
      <TableContainer component={Paper} sx={{ mb: 3 }} elevation={1}>
        <Table size="small" aria-label="KPI by calling domain">
          <TableHead>
            <TableRow>
              <TableCell>Calling domain</TableCell>
              <TableCell align="right">Volume</TableCell>
              <TableCell align="right">Σ estimated</TableCell>
              <TableCell align="right">Σ actual</TableCell>
              <TableCell align="right">Success %</TableCell>
            </TableRow>
          </TableHead>
          <TableBody>
            {(data.by_calling_domain ?? []).map((row) => (
              <TableRow key={row.calling_domain}>
                <TableCell>{row.calling_domain}</TableCell>
                <TableCell align="right">{row.volume}</TableCell>
                <TableCell align="right">
                  {formatCost(row.total_estimated_cost)}
                </TableCell>
                <TableCell align="right">
                  {formatCost(row.total_actual_cost)}
                </TableCell>
                <TableCell align="right">
                  {formatPct(row.terminal_success_rate)}
                </TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      </TableContainer>

      <Typography variant="h6" gutterBottom>
        By country
      </Typography>
      <TableContainer component={Paper} elevation={1}>
        <Table size="small" aria-label="KPI by country">
          <TableHead>
            <TableRow>
              <TableCell>Country</TableCell>
              <TableCell align="right">Volume</TableCell>
              <TableCell align="right">Σ estimated</TableCell>
              <TableCell align="right">Σ actual</TableCell>
              <TableCell align="right">Success %</TableCell>
            </TableRow>
          </TableHead>
          <TableBody>
            {data.by_country.map((row) => (
              <TableRow key={row.country_code}>
                <TableCell>{row.country_code}</TableCell>
                <TableCell align="right">{row.volume}</TableCell>
                <TableCell align="right">
                  {formatCost(row.total_estimated_cost)}
                </TableCell>
                <TableCell align="right">
                  {formatCost(row.total_actual_cost)}
                </TableCell>
                <TableCell align="right">
                  {formatPct(row.terminal_success_rate)}
                </TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      </TableContainer>
    </Box>
  );
};
