import { Container, Typography } from '@mui/material';

import { MembershipSmsFeature } from '../../features/membership-sms';

export const MembershipRegistrationPage = () => (
  <Container maxWidth="md" component="main">
    <Typography variant="h4" component="h1" gutterBottom>
      Membership Registration
    </Typography>
    <MembershipSmsFeature />
  </Container>
);
