import type { ChangeEvent } from 'react';

import { FormControl, InputLabel, Select } from '@mui/material';

import type { SmsScenarioId } from './types';

const OPTIONS: readonly { readonly id: SmsScenarioId; readonly label: string }[] = [
  { id: 'registration-otp', label: 'Registration Success + OTP' },
];

interface SmsTypeSelectorProps {
  readonly value: SmsScenarioId;
  readonly onChange: (next: SmsScenarioId) => void;
  readonly disabled?: boolean;
}

export const SmsTypeSelector = ({
  value,
  onChange,
  disabled = false,
}: SmsTypeSelectorProps) => {
  const handleChange = (event: ChangeEvent<HTMLSelectElement>): void => {
    onChange(event.target.value as SmsScenarioId);
  };

  return (
    <FormControl fullWidth size="small" disabled={disabled}>
      <InputLabel id="sms-type-select-label" htmlFor="sms-type-select">
        SMS Type
      </InputLabel>
      <Select
        labelId="sms-type-select-label"
        id="sms-type-select"
        name="smsType"
        label="SMS Type"
        native
        value={value}
        onChange={handleChange}
        inputProps={{ 'aria-label': 'SMS Type' }}
      >
        {OPTIONS.map((option) => (
          <option key={option.id} value={option.id}>
            {option.label}
          </option>
        ))}
      </Select>
    </FormControl>
  );
};
