import { Box } from '@mui/material';
import { styled } from '@mui/material/styles';
import { visuallyHidden } from '@mui/utils';

interface OtpDisplayProps {
  readonly otp: string;
}

const OtpDigit = styled('span')(({ theme }) => ({
  display: 'inline-flex',
  alignItems: 'center',
  justifyContent: 'center',
  minWidth: '2.75rem',
  padding: '0.5rem 0.65rem',
  fontFamily:
    'ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, "Liberation Mono", "Courier New", monospace',
  fontSize: '2.25rem',
  fontWeight: 600,
  lineHeight: 1,
  color: theme.palette.text.primary,
  background: `linear-gradient(180deg, ${theme.palette.grey[50]} 0%, ${theme.palette.grey[300]} 100%)`,
  border: `1px solid ${theme.palette.grey[400]}`,
  borderRadius: theme.shape.borderRadius * 1.5,
  boxShadow: `0 1px 2px rgba(15,23,42,0.08), inset 0 1px 0 rgba(255,255,255,0.8)`,
}));

export const OtpDisplay = ({ otp }: OtpDisplayProps) => {
  const digits = otp.split('');

  return (
    <Box
      sx={{
        display: 'flex',
        flexDirection: 'column',
        alignItems: 'center',
      }}
      role="status"
      aria-label="OTP code"
    >
      <Box component="span" sx={visuallyHidden}>
        {otp}
      </Box>
      <Box
        sx={{
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          gap: '0.5rem',
        }}
        aria-hidden="true"
      >
        {digits.map((digit, index) => (
          <OtpDigit key={`${digit}-${String(index)}`}>{digit}</OtpDigit>
        ))}
      </Box>
    </Box>
  );
};
