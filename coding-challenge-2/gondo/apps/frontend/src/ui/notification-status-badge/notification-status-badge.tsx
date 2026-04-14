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
  if (variant === 'neutral') {
    return 'default';
  }
  if (variant === 'success') {
    return 'success';
  }
  if (variant === 'error') {
    return 'error';
  }
  return 'warning';
};

export const NotificationStatusBadge = ({
  statusLabel,
}: NotificationStatusBadgeProps) => {
  const variant = resolveVariant(statusLabel);

  return (
    <Chip
      size="small"
      label={statusLabel}
      color={variantToChipColor(variant)}
      variant="outlined"
    />
  );
};
