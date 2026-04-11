import type { FormEvent, ChangeEvent } from 'react';

import {
  Button,
  FormControl,
  InputLabel,
  Select,
  Stack,
  TextField,
} from '@mui/material';

import type { MembershipSmsFormValues } from './types';
import {
  getE164PlaceholderForCountry,
  MEMBERSHIP_SMS_COUNTRY_OPTIONS,
} from './build-phone-number';

interface MembershipSmsFormProps {
  readonly values: MembershipSmsFormValues;
  readonly onChange: (patch: Partial<MembershipSmsFormValues>) => void;
  readonly onSubmit: () => void;
  readonly disabled?: boolean;
}

export const MembershipSmsForm = ({
  values,
  onChange,
  onSubmit,
  disabled = false,
}: MembershipSmsFormProps) => {
  const phonePlaceholder = getE164PlaceholderForCountry(values.countryCode);

  const handleSubmit = (event: FormEvent<HTMLFormElement>): void => {
    event.preventDefault();
    onSubmit();
  };

  const handleCountryChange = (event: ChangeEvent<HTMLSelectElement>): void => {
    onChange({ countryCode: event.target.value });
  };

  return (
    <form onSubmit={handleSubmit}>
      <Stack spacing={2}>
        <TextField
          id="membership-message-id"
          name="messageId"
          label="Message ID"
          value={values.messageId}
          disabled={disabled}
          onChange={(event: ChangeEvent<HTMLInputElement>) => {
            onChange({ messageId: event.target.value });
          }}
        />
        <TextField
          id="membership-recipient"
          name="recipient"
          label="Recipient"
          value={values.recipient}
          disabled={disabled}
          onChange={(event: ChangeEvent<HTMLInputElement>) => {
            onChange({ recipient: event.target.value });
          }}
        />
        <FormControl fullWidth size="small" disabled={disabled}>
          <InputLabel
            id="membership-country-code-label"
            htmlFor="membership-country-code"
          >
            Country Code
          </InputLabel>
          <Select
            labelId="membership-country-code-label"
            id="membership-country-code"
            name="countryCode"
            label="Country Code"
            native
            value={values.countryCode}
            onChange={handleCountryChange}
          >
            {MEMBERSHIP_SMS_COUNTRY_OPTIONS.map((option) => (
              <option key={option.value} value={option.value}>
                {option.label}
              </option>
            ))}
          </Select>
        </FormControl>
        <TextField
          id="membership-phone-number"
          name="phoneNumber"
          label="Phone Number"
          value={values.phoneNumber}
          placeholder={phonePlaceholder}
          disabled={disabled}
          onChange={(event: ChangeEvent<HTMLInputElement>) => {
            onChange({ phoneNumber: event.target.value });
          }}
          helperText={
            <>
              Use international format (e.g. {phonePlaceholder}) or national digits; the app adds the
              correct + prefix for the selected country when needed.
            </>
          }
          slotProps={{
            formHelperText: { id: 'membership-phone-hint' },
            htmlInput: { 'aria-describedby': 'membership-phone-hint' },
          }}
        />
        <TextField
          id="membership-message-content"
          name="content"
          label="Message Content"
          value={values.content}
          disabled={disabled}
          multiline
          rows={4}
          onChange={(
            event: ChangeEvent<HTMLInputElement | HTMLTextAreaElement>,
          ) => {
            onChange({ content: event.target.value });
          }}
        />
        <Button
          type="submit"
          variant="contained"
          disabled={disabled}
          sx={{ alignSelf: 'center' }}
        >
          Send SMS
        </Button>
      </Stack>
    </form>
  );
};
