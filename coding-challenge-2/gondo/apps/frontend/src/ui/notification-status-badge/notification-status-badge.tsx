import { Chip } from '@mui/material';

interface NotificationStatusBadgeProps {
  readonly statusLabel: string;
}

type BadgeVariant = 'success' | 'error' | 'warning' | 'neutral';

const resolveVariant = (statusLabel: string): BadgeVariant => {
  if (statusLabel === 'Send-success') {
    return 'success';
  }
  if (statusLabel === 'Send-failed') {
    return 'error';
  }
  if (statusLabel === 'Carrier-rejected') {
    return 'warning';
  }
  if (
    statusLabel === 'New' ||
    statusLabel === 'Send-to-provider' ||
    statusLabel === 'Queue' ||
    statusLabel === 'Send-to-carrier'
  ) {
    return 'neutral';
  }
  return 'neutral';
};

const variantToChipColor = (
  variant: BadgeVariant,
): 'success' | 'error' | 'warning' | 'default' | 'info' => {
  switch (variant) {
    case 'neutral':
      return 'default';
    case 'success':
      return 'success';
    case 'error':
      return 'error';
    case 'warning':
    default:
      return 'warning';
  }
};

const getLabel = (statusLabel: string): string => {
  switch (statusLabel) {
    case 'Send-success':
      return 'Success';
    case 'Send-failed':
      return 'Failed';
    case 'Carrier-rejected':
      return 'Carrier Rejected';
    case 'New':
      return 'New';
    case 'Send-to-provider':
      return 'Provider Received';
    case 'Queue':
      return 'Queued';
    case 'Send-to-carrier':
      return 'Carrier Received';
    default:
      return 'Unknown';
  }
};

export const NotificationStatusBadge = ({
  statusLabel,
}: NotificationStatusBadgeProps) => {
  const variant = resolveVariant(statusLabel);

  return (
    <Chip
      size="small"
      label={getLabel(statusLabel)}
      color={variantToChipColor(variant)}
      variant="outlined"
    />
  );
};
