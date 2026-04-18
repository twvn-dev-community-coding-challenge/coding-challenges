export { generateSixDigitOtp } from './otp';
export {
  createNotificationApi,
} from './notification-api';
export type {
  ApiErrorBody,
  ChannelType,
  CreateNotificationRequest,
  DispatchRequest,
  ListNotificationsData,
  NotificationResource,
  Result,
  RetryRequest,
  SmsChannelPayload,
} from './notification-api';
/** Generated from `docs/openapi/notification-service.openapi.json` — run `yarn nx run ts-core:generate-openapi-types`. */
export type {
  components,
  operations,
  paths,
  webhooks,
} from './notification-api/openapi.generated';
