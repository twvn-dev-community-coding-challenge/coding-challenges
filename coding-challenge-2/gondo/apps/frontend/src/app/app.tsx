import { Box, CssBaseline, ThemeProvider } from '@mui/material';
import {
  BrowserRouter,
  Navigate,
  Route,
  Routes,
} from 'react-router';

import { CountdownProvider } from '../context';
import { MembershipRegistrationPage } from '../pages/membership-registration';
import { NotificationTrackingPage } from '../pages/notification-tracking';
import { theme } from '../theme';
import { NavBar } from '../ui/nav-bar/nav-bar';

const navLinks = [
  { to: '/membership', label: 'Membership Registration' },
  { to: '/tracking', label: 'Notification Tracking' },
] as const;

export const AppRoutes = () => (
  <Box
    sx={{
      minHeight: '100vh',
      display: 'flex',
      flexDirection: 'column',
      bgcolor: 'background.default',
    }}
  >
    <NavBar links={navLinks} />
    <Box sx={{ flex: 1, width: '100%', py: 3 }}>
      <Routes>
        <Route path="/" element={<Navigate to="/membership" replace />} />
        <Route
          path="/membership"
          element={
            <Box sx={{ maxWidth: 960, mx: 'auto', px: { xs: 2, sm: 3 } }}>
              <MembershipRegistrationPage />
            </Box>
          }
        />
        <Route
          path="/tracking"
          element={
            <Box sx={{ px: { xs: 2, sm: 3 } }}>
              <NotificationTrackingPage />
            </Box>
          }
        />
      </Routes>
    </Box>
  </Box>
);

export const App = () => (
  <BrowserRouter>
    <ThemeProvider theme={theme}>
      <CssBaseline />
      <CountdownProvider>
        <AppRoutes />
      </CountdownProvider>
    </ThemeProvider>
  </BrowserRouter>
);

export default App;
