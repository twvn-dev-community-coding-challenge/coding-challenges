package com.ridehailing.dispatch.adapters.messaging;

import com.ridehailing.common.events.InboundEvents;
import com.ridehailing.contracts.Topics;
import com.ridehailing.dispatch.application.DispatchEventProcessor;
import lombok.RequiredArgsConstructor;
import org.apache.kafka.clients.consumer.ConsumerRecord;
import org.springframework.kafka.annotation.KafkaListener;
import org.springframework.stereotype.Component;

/**
 * Dispatch's subscriptions (consumer group {@code dispatch}): ride-events (a new ride, a ride ending), its own
 * dispatch-events (rematch after a rejection or expiry), and driver-events (supply copy). Failures roll back and are
 * retried by the container's error handler, then dead-lettered (see MessagingConfig).
 */
@Component
@RequiredArgsConstructor
public class DispatchEventListener {

    private final InboundEvents inbound;
    private final DispatchEventProcessor handler;

    @KafkaListener(id = "dispatch", topics = {Topics.RIDE_EVENTS, Topics.DISPATCH_EVENTS, Topics.DRIVER_EVENTS}, groupId = DispatchEventProcessor.CONSUMER)
    public void onEvent(ConsumerRecord<String, String> record) {
        inbound.handle(DispatchEventProcessor.CONSUMER, record, DispatchEventProcessor.HANDLED_EVENTS, handler::process);
    }
}
