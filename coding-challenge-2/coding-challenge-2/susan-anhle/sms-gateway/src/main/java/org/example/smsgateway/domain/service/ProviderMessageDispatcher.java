package org.example.smsgateway.domain.service;

import org.example.smsgateway.domain.model.agreement.Provider;
import org.example.smsgateway.domain.model.message.Message;

import java.util.Map;

public class ProviderMessageDispatcher {
    private final Map<
            Provider, ProviderSendMessageGateway> handlers;

    public ProviderMessageDispatcher(Map<Provider, ProviderSendMessageGateway> handlers) {
        this.handlers = handlers;
    }

    public void dispatch(Message message, Provider provider) {
        var handler = handlers.get(provider);
        handler.handle(message);
    }

    public interface ProviderSendMessageGateway {
        void handle(Message message);
    }
}
