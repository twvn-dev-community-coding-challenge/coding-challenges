import { ThemeProvider } from '@mui/material';
import { render, type RenderOptions } from '@testing-library/react';
import type { ReactElement, ReactNode } from 'react';

import { CountdownProvider } from './context';
import { theme } from './theme';

const WithTheme = ({ children }: { readonly children: ReactNode }) => (
  <ThemeProvider theme={theme}>
    <CountdownProvider>{children}</CountdownProvider>
  </ThemeProvider>
);

export const renderWithTheme = (
  ui: ReactElement,
  options?: Omit<RenderOptions, 'wrapper'>,
): ReturnType<typeof render> =>
  render(ui, { wrapper: WithTheme, ...options });
