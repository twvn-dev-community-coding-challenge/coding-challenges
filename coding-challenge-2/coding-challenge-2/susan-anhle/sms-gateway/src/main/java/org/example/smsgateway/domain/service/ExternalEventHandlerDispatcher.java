package org.example.smsgateway.domain.service;

import org.example.smsgateway.application.handlers.callbackHandler.ExternalEventHandler;
import org.example.smsgateway.domain.model.common.ExternalEvent;
import org.example.smsgateway.domain.model.message.Message;

import java.util.Map;

public class ExternalEventHandlerDispatcher {
    private final Map<
            Class<? extends ExternalEvent>, ExternalEventHandler<? extends ExternalEvent>> handlers;

    public ExternalEventHandlerDispatcher(Map<Class<? extends ExternalEvent>, ExternalEventHandler<? extends ExternalEvent>> handlers) {
        this.handlers = handlers;
    }


    public void dispatch(ExternalEvent externalEvent) {
        Class<? extends ExternalEvent> externalEventType = externalEvent.getClass();
        ExternalEventHandler handler = handlers.get(externalEventType);
        if (handler == null) {
            throw new IllegalStateException("Need to define a handler for this external event: " + externalEventType);
        }
        handler.handle(externalEvent);
    }
}
