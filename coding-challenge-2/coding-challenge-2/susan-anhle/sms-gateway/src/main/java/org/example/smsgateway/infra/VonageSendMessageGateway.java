package org.example.smsgateway.infra;

import org.example.smsgateway.domain.model.message.Message;
import org.example.smsgateway.domain.service.ProviderMessageDispatcher;

public class VonageSendMessageGateway implements ProviderMessageDispatcher.ProviderSendMessageGateway {
    @Override
    public void handle(Message message) {
        // send to rest api: https://vonage.com/api/sendMessage
    }
}
