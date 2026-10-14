package org.example.smsgateway.application.handlers.callbackHandler;

import org.example.smsgateway.domain.model.common.ExternalEvent;

public interface ExternalEventHandler<TEventType extends ExternalEvent> {
    void handle(TEventType event);
}
