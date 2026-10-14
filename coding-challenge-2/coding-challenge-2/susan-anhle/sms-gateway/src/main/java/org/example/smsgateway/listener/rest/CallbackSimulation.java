package org.example.smsgateway.listener.rest;

import org.example.smsgateway.domain.model.common.ExternalEvent;
import org.example.smsgateway.domain.model.externalevent.CarrierRejected;
import org.example.smsgateway.domain.model.externalevent.Queued;
import org.example.smsgateway.domain.model.externalevent.SentSuccessfully;
import org.example.smsgateway.domain.model.externalevent.SentToCarrier;
import org.example.smsgateway.domain.service.ExternalEventHandlerDispatcher;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.RestController;

import java.math.BigDecimal;

@RestController
public class CallbackSimulation {

    private static final String QUEUE_STATE = "queue";
    private static final String SENT_TO_CARRIER_STATE = "send-to-carrier";
    private static final String SENT_SUCCESSFULLY_STATE = "send-success";
    private static final String CARRIER_REJECTED_STATE = "carrier-rejected";
    private final ExternalEventHandlerDispatcher eventHandlerDispatcher;

    public CallbackSimulation(ExternalEventHandlerDispatcher eventHandlerDispatcher) {
        this.eventHandlerDispatcher = eventHandlerDispatcher;
    }

    @GetMapping("/api/v1/callback-simulation")
    public void callBackSimulation(CallbackMessage message) {
        eventHandlerDispatcher.dispatch(translate(message));
    }

    private ExternalEvent translate(CallbackMessage message) {
        return switch (message.state.toLowerCase()) {
            case QUEUE_STATE -> new Queued(message.messageId);
            case SENT_TO_CARRIER_STATE -> new SentToCarrier(message.messageId);
            case SENT_SUCCESSFULLY_STATE -> new SentSuccessfully(message.messageId, message.actualCost);
            case CARRIER_REJECTED_STATE -> new CarrierRejected(message.messageId);
            default -> throw new UnsupportedOperationException("Operation not supported");
        };
    }

    public record CallbackMessage(
            String messageId, String state, BigDecimal actualCost
    ) {
    }
}
