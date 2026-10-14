/**
 * Represents all possible states in the SMS delivery lifecycle.
 */
export enum SmsStatus {
  NEW = "New",
  SEND_TO_PROVIDER = "Send-to-provider",
  QUEUE = "Queue",
  SEND_TO_CARRIER = "Send-to-carrier",
  SEND_SUCCESS = "Send-success",
  SEND_FAILED = "Send-failed",
  CARRIER_REJECTED = "Carrier-rejected",
}
