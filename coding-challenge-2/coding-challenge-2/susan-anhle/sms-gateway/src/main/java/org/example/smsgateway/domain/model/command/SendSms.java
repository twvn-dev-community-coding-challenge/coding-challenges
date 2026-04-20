package org.example.smsgateway.domain.model.command;

public class SendSms implements Command {
    private final String messageId;
    private final String message;
    private final String country;
    private final String phoneNumber;
    public String getMessageId() {
        return messageId;
    }

    public String getMessage() {
        return message;
    }

    public String getCountry() {
        return country;
    }

    public String getPhoneNumber() {
        return phoneNumber;
    }


    public SendSms(String messageId, String message, String country, String phoneNumber) {
        ensureAttributeAvailable(messageId, message, country, phoneNumber);
        ensureMessageLength(message);
        this.messageId = messageId;
        this.message = message;
        this.country = country;
        this.phoneNumber = phoneNumber;
    }

    private void ensureAttributeAvailable(String messageId, String message, String country, String phoneNumber) {
        if (messageId == null || message == null || country == null || phoneNumber == null || messageId.isEmpty() || message.isEmpty()
                || country.isEmpty() || phoneNumber.isEmpty()) {
            throw new IllegalArgumentException("All attributes must be provided and not empty");
        }
    }

    private static void ensureMessageLength(String message) {
        if (message.length() > 999) {
            throw new IllegalArgumentException("Message length maximum is 999");
        }
    }
}
