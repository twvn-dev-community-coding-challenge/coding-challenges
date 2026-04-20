import { screen, waitFor } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';

import { renderWithTheme } from '../../test-utils';
import { SmsKpisPage } from './sms-kpis-page';

const mockGetSmsKpis = vi.fn();

vi.mock('@gondo/ts-core', () => ({
  createNotificationApi: () => ({
    getSmsKpis: mockGetSmsKpis,
  }),
}));

const sampleKpis = {
  source: 'in_memory_notifications',
  created_at_filter: { from: null, to: null },
  currency_note: 'note',
  overall: {
    volume: 2,
    total_notifications: 2,
    total_estimated_cost: 0.04,
    total_actual_cost: 0.03,
    with_estimate_count: 2,
    with_actual_count: 1,
    send_success: 1,
    send_failed: 0,
    carrier_rejected: 0,
    in_flight: 1,
    terminal_failure: 0,
    terminal_success_rate: 1,
    terminal_failure_rate: 0,
  },
  by_calling_domain: [],
  by_provider: [
    {
      provider_id: 'prv_01',
      volume: 2,
      total_estimated_cost: 0.04,
      total_actual_cost: 0.03,
      with_estimate_count: 2,
      with_actual_count: 1,
      send_success: 1,
      send_failed: 0,
      carrier_rejected: 0,
      in_flight: 1,
      terminal_failure: 0,
      terminal_success_rate: 1,
      terminal_failure_rate: 0,
    },
  ],
  by_country: [
    {
      country_code: 'VN',
      volume: 2,
      total_estimated_cost: 0.04,
      total_actual_cost: 0.03,
      with_estimate_count: 2,
      with_actual_count: 1,
      send_success: 1,
      send_failed: 0,
      carrier_rejected: 0,
      in_flight: 1,
      terminal_failure: 0,
      terminal_success_rate: 1,
      terminal_failure_rate: 0,
    },
  ],
};

describe('SmsKpisPage', () => {
  it('loads and shows KPI tables', async () => {
    mockGetSmsKpis.mockResolvedValue({ ok: true, value: sampleKpis });

    renderWithTheme(<SmsKpisPage />);

    expect(await screen.findByRole('heading', { name: /sms observability/i })).toBeTruthy();
    await waitFor(() =>
      expect(screen.queryByTestId('kpis-loading')).toBeNull(),
    );
    expect(screen.getByText(/overall/i)).toBeTruthy();
    expect(screen.getByRole('table', { name: /kpi by provider/i })).toBeTruthy();
    expect(screen.getByRole('table', { name: /kpi by country/i })).toBeTruthy();
    expect(screen.getByText('prv_01')).toBeTruthy();
    expect(screen.getByText('VN')).toBeTruthy();
  });

  it('shows error when API fails', async () => {
    mockGetSmsKpis.mockResolvedValue({
      ok: false,
      error: { code: 'HTTP_ERROR', message: 'down' },
    });

    renderWithTheme(<SmsKpisPage />);

    expect(await screen.findByText(/failed to load kpis/i)).toBeTruthy();
  });
});
