import { useCallback, useMemo, useReducer } from 'react';

import {
  Alert,
  Box,
  Button,
  Link,
  Paper,
  Stack,
  Typography,
} from '@mui/material';
import { Link as RouterLink } from 'react-router';
import { createNotificationApi } from '@gondo/ts-core';

import { DEFAULT_COUNTDOWN_DURATION, useCountdown } from '../../context';
import { OtpDisplay } from '../../ui/otp-display/otp-display';
import { buildSmsPhoneNumber } from './build-phone-number';
import { MembershipSmsForm } from './membership-sms-form';
import { SmsTypeSelector } from './sms-type-selector';
import type { ClientPhase, MembershipSmsFormValues, SmsScenarioId } from './types';

const DEFAULT_MESSAGE_TEMPLATE =
  'Your membership registration is successful. Your OTP code is: {{OTP}}. Valid for 2 minutes.';

const createInitialFormValues = (): MembershipSmsFormValues => ({
  messageId: globalThis.crypto.randomUUID(),
  recipient: '',
  countryCode: 'VN',
  phoneNumber: '',
  content: DEFAULT_MESSAGE_TEMPLATE,
});

interface FeatureState {
  readonly phase: ClientPhase;
  readonly formValues: MembershipSmsFormValues;
  readonly scenario: SmsScenarioId;
  readonly otp: string | null;
  readonly notificationId: string | null;
  readonly error: string | null;
}

type Action =
  | { readonly type: 'update_form'; readonly patch: Partial<MembershipSmsFormValues> }
  | { readonly type: 'set_scenario'; readonly scenario: SmsScenarioId }
  | { readonly type: 'submit_start' }
  | {
      readonly type: 'submit_success';
      readonly payload: {
        readonly notificationId: string;
        readonly otp: string;
      };
    }
  | { readonly type: 'submit_fail'; readonly message: string }
  | { readonly type: 'new_sms' };

const buildInitialState = (): FeatureState => ({
  phase: 'idle',
  formValues: createInitialFormValues(),
  scenario: 'registration-otp',
  otp: null,
  notificationId: null,
  error: null,
});

const reducer = (state: FeatureState, action: Action): FeatureState => {
  switch (action.type) {
    case 'update_form':
      return {
        ...state,
        formValues: { ...state.formValues, ...action.patch },
      };
    case 'set_scenario':
      return { ...state, scenario: action.scenario };
    case 'submit_start':
      return { ...state, phase: 'submitting', error: null };
    case 'submit_success':
      return {
        ...state,
        phase: 'sent',
        notificationId: action.payload.notificationId,
        otp: action.payload.otp,
        error: null,
      };
    case 'submit_fail':
      return { ...state, phase: 'idle', error: action.message };
    case 'new_sms':
      return buildInitialState();
  }
};

interface MembershipSmsFeatureProps {
  readonly countdownDurationSeconds?: number;
}

export const MembershipSmsFeature = ({
  countdownDurationSeconds = DEFAULT_COUNTDOWN_DURATION,
}: MembershipSmsFeatureProps) => {
  const [state, dispatch] = useReducer(reducer, undefined, buildInitialState);
  const { startCountdown } = useCountdown();

  const api = useMemo(() => createNotificationApi(), []);

  const handleFormChange = useCallback((patch: Partial<MembershipSmsFormValues>) => {
    dispatch({ type: 'update_form', patch });
  }, []);

  const handleScenarioChange = useCallback((scenario: SmsScenarioId) => {
    dispatch({ type: 'set_scenario', scenario });
  }, []);

  const handleSubmit = useCallback(async () => {
    if (state.phase !== 'idle') {
      return;
    }

    if (!state.formValues.phoneNumber.trim()) {
      dispatch({
        type: 'submit_fail',
        message: 'Phone number is required.',
      });
      return;
    }

    dispatch({ type: 'submit_start' });

    const createResult = await api.createNotification({
      message_id: state.formValues.messageId,
      recipient: state.formValues.recipient,
      content: state.formValues.content,
      channel_payload: {
        country_code: state.formValues.countryCode,
        phone_number: buildSmsPhoneNumber(
          state.formValues.countryCode,
          state.formValues.phoneNumber,
        ),
      },
      issue_server_otp: true,
    });

    if (!createResult.ok) {
      dispatch({ type: 'submit_fail', message: createResult.error.message });
      return;
    }

    const notificationId = createResult.value.notification_id;
    const otp =
      createResult.value.otp_plaintext ??
      (createResult.value.content.includes('OTP:')
        ? createResult.value.content.split('OTP:').pop()?.trim() ?? ''
        : '');

    const dispatchResult = await api.dispatchNotification(notificationId, {
      as_of: new Date().toISOString(),
    });

    if (!dispatchResult.ok) {
      dispatch({ type: 'submit_fail', message: dispatchResult.error.message });
      return;
    }

    const createdAtMs = new Date(createResult.value.created_at).getTime();
    startCountdown({
      notificationId,
      otp,
      startedAt: Number.isNaN(createdAtMs) ? Date.now() : createdAtMs,
      durationSeconds: countdownDurationSeconds,
    });

    dispatch({
      type: 'submit_success',
      payload: {
        notificationId,
        otp,
      },
    });
  }, [api, countdownDurationSeconds, startCountdown, state.formValues, state.phase]);

  const handleNewSms = useCallback(() => {
    dispatch({ type: 'new_sms' });
  }, []);

  const formDisabled = state.phase !== 'idle';
  const selectorDisabled = state.phase !== 'idle';
  const showSuccess = state.phase === 'sent' && state.otp !== null;

  return (
    <Stack spacing={3}>
      <Paper
        variant="outlined"
        sx={{
          p: 2,
          borderColor: 'grey.300',
        }}
      >
        <Typography variant="h6" component="h2" gutterBottom>
          SMS scenario
        </Typography>
        <SmsTypeSelector
          value={state.scenario}
          onChange={handleScenarioChange}
          disabled={selectorDisabled}
        />
      </Paper>

      <Paper
        variant="outlined"
        sx={{
          p: 2,
          borderColor: 'grey.300',
        }}
      >
        <Typography variant="h6" component="h2" gutterBottom>
          Compose SMS
        </Typography>
        <MembershipSmsForm
          values={state.formValues}
          onChange={handleFormChange}
          onSubmit={handleSubmit}
          disabled={formDisabled}
        />
      </Paper>

      {state.error !== null && state.error !== '' && (
        <Alert severity="error" role="alert">
          {state.error}
        </Alert>
      )}

      {showSuccess && (
        <Paper
          variant="outlined"
          sx={{
            p: 2,
            borderColor: 'grey.400',
            bgcolor: 'background.paper',
          }}
          aria-live="polite"
        >
          <Typography variant="h6" component="h2" gutterBottom>
            Delivery
          </Typography>
          <Stack spacing={2}>
            <Alert severity="success">
              SMS dispatched! Track it on the{' '}
              <Link component={RouterLink} to="/tracking" underline="always">
                Notification Tracking
              </Link>{' '}
              page.
            </Alert>
            <Box
              sx={{
                display: 'flex',
                alignItems: 'center',
                gap: 2,
                flexWrap: 'wrap',
              }}
            >
              <Typography
                component="span"
                sx={{ color: 'text.secondary', fontWeight: 600 }}
              >
                OTP
              </Typography>
              <OtpDisplay otp={state.otp} />
            </Box>
            <Button
              variant="outlined"
              type="button"
              onClick={handleNewSms}
              sx={{ alignSelf: 'center' }}
            >
              New SMS
            </Button>
          </Stack>
        </Paper>
      )}
    </Stack>
  );
};
