package org.example.smsgateway.domain.model.view;

public class SMSVolumePerProvider {
    private String providerId;
    private long messageCount;

    public SMSVolumePerProvider(String providerId, long messageCount) {
        this.providerId = providerId;
        this.messageCount = messageCount;
    }

    public void incrementMessageCount() {
        this.messageCount++;
    }

    public String getProviderId() {
        return providerId;
    }

    public long getMessageCount() {
        return messageCount;
    }
}
