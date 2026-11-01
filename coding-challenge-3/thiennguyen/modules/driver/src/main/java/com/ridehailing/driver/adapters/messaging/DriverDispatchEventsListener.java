package com.ridehailing.driver.adapters.messaging;

import com.ridehailing.common.events.InboundEvents;
import com.ridehailing.contracts.Topics;
import com.ridehailing.driver.application.AssignmentProjection;
import lombok.RequiredArgsConstructor;
import org.apache.kafka.clients.consumer.ConsumerRecord;
import org.springframework.kafka.annotation.KafkaListener;
import org.springframework.stereotype.Component;

/**
 * Driver's subscription to dispatch-events (consumer group {@code driver}); only assignment changes matter here.
 */
@Component
@RequiredArgsConstructor
public class DriverDispatchEventsListener {

    private final InboundEvents inbound;
    private final AssignmentProjection handler;

    @KafkaListener(id = "driver-projection", topics = Topics.DISPATCH_EVENTS, groupId = AssignmentProjection.CONSUMER)
    public void onEvent(ConsumerRecord<String, String> record) {
        inbound.handle(AssignmentProjection.CONSUMER, record, AssignmentProjection.HANDLED_EVENTS, (type, envelope) -> handler.apply(envelope));
    }
}
