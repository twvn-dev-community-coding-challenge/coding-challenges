import RefreshIcon from '@mui/icons-material/Refresh';
import {
  Alert,
  Box,
  Button,
  Card,
  CardContent,
  Chip,
  CircularProgress,
  Typography,
} from '@mui/material';
import { useCallback, useEffect, useMemo, useState } from 'react';

import { createNotificationApi } from '@gondo/ts-core';
import type { MockScenario, MockScenariosData } from '@gondo/ts-core';

const Outcomes = [
  'send_success',
  'send_failed',
  'carrier_rejected',
  'retry_then_success',
] as const;

type Outcome = (typeof Outcomes)[number];

const formatCost = (n: number): string => n.toFixed(4);

const formatOptionalCost = (value: number | null | undefined): string => {
  if (value === null || value === undefined) {
    return '—';
  }
  return formatCost(value);
};

const getOutcomeChipColor = (
  outcome: string,
): 'success' | 'error' | 'warning' | 'info' => {
  if (outcome === 'send_success') {
    return 'success';
  }
  if (outcome === 'send_failed') {
    return 'error';
  }
  if (outcome === 'carrier_rejected') {
    return 'warning';
  }
  return 'info';
};

const ScenarioCard = ({ scenario }: { readonly scenario: MockScenario }) => (
  <Card
    elevation={2}
    sx={{ height: '100%', display: 'flex', flexDirection: 'column' }}
  >
    <CardContent sx={{ flex: 1 }}>
      <Box
        sx={{
          display: 'flex',
          justifyContent: 'space-between',
          alignItems: 'flex-start',
          mb: 1,
          gap: 1,
        }}
      >
        <Typography variant="subtitle1" fontWeight={600}>
          {scenario.phone_number}
        </Typography>
        <Box
          sx={{
            display: 'flex',
            flexWrap: 'wrap',
            gap: 0.5,
            justifyContent: 'flex-end',
          }}
        >
          <Chip
            label={scenario.country_code}
            size="small"
            variant="outlined"
          />
          <Chip
            label={scenario.outcome}
            size="small"
            color={getOutcomeChipColor(scenario.outcome)}
          />
        </Box>
      </Box>
      <Typography variant="body2" color="text.secondary" sx={{ mb: 1.5 }}>
        {scenario.description}
      </Typography>
      <Typography variant="body2" sx={{ mb: 0.5 }}>
        <strong>Carrier:</strong> {scenario.carrier}
      </Typography>
      <Typography variant="body2" sx={{ mb: 0.5 }}>
        <strong>Provider:</strong> {scenario.expected_provider}
      </Typography>
      <Typography variant="body2" sx={{ mb: 0.5 }}>
        <strong>User story:</strong> {scenario.user_story}
      </Typography>
      <Typography variant="body2" sx={{ mb: 0.5 }}>
        <strong>Calling domain:</strong> {scenario.calling_domain ?? '—'}
      </Typography>
      <Typography variant="body2" sx={{ mb: 0.5 }}>
        <strong>Estimated cost:</strong>{' '}
        {formatOptionalCost(scenario.estimated_cost)}
      </Typography>
      <Typography variant="body2">
        <strong>Actual cost:</strong> {formatOptionalCost(scenario.actual_cost)}
      </Typography>
    </CardContent>
  </Card>
);

export const MockedScenariosPage = () => {
  const [data, setData] = useState<MockScenariosData | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);

  const load = useCallback(async () => {
    setLoading(true);
    setError(null);
    const api = createNotificationApi();
    const res = await api.getMockScenarios();
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

  const outcomeCounts = useMemo(() => {
    if (data === null) {
      return null;
    }
    const initial: Record<Outcome, number> = {
      send_success: 0,
      send_failed: 0,
      carrier_rejected: 0,
      retry_then_success: 0,
    };
    return data.scenarios.reduce<Record<Outcome, number>>((acc, s) => {
      const key = s.outcome as Outcome;
      if (key in acc) {
        acc[key] += 1;
      }
      return acc;
    }, initial);
  }, [data]);

  if (loading) {
    return (
      <Box
        data-testid="mock-scenarios-loading"
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
          Mocked Scenarios
        </Typography>
        <Alert severity="error">Failed to load mock scenarios: {error}</Alert>
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

  if (data === null || outcomeCounts === null) {
    return null;
  }

  const summaryParts = Outcomes.map(
    (o) => `${o}: ${String(outcomeCounts[o])}`,
  ).join(' · ');

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
          Mocked Scenarios
        </Typography>
        <Button
          startIcon={<RefreshIcon />}
          onClick={() => void load()}
          variant="outlined"
        >
          Refresh
        </Button>
      </Box>
      <Typography variant="body1" color="text.secondary" sx={{ mb: 3 }}>
        {data.scenarios.length} scenarios · {summaryParts}
      </Typography>
      <Box
        sx={{
          display: 'flex',
          flexWrap: 'wrap',
          gap: 2,
        }}
      >
        {data.scenarios.map((scenario) => (
          <Box
            key={scenario.phone_number}
            sx={{
              flex: '1 1 320px',
              minWidth: { xs: '100%', sm: 280 },
              maxWidth: {
                sm: 'calc(50% - 8px)',
                md: 'calc(33.333% - 11px)',
              },
            }}
          >
            <ScenarioCard scenario={scenario} />
          </Box>
        ))}
      </Box>
    </Box>
  );
};
