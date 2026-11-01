package com.ridehailing.ride.adapters.messaging;

import com.ridehailing.common.events.InboundEvents;
import com.ridehailing.contracts.Topics;
import com.ridehailing.ride.application.DispatchProjection;
import lombok.RequiredArgsConstructor;
import org.apache.kafka.clients.consumer.ConsumerRecord;
import org.springframework.kafka.annotation.KafkaListener;
import org.springframework.stereotype.Component;

/**
 * Ride's subscription to dispatch-events (consumer group {@code ride}).
 */
@Component
@RequiredArgsConstructor
public class RideDispatchEventsListener {

    private final InboundEvents inbound;
    private final DispatchProjection handler;

    @KafkaListener(id = "ride-projection", topics = Topics.DISPATCH_EVENTS, groupId = DispatchProjection.CONSUMER)
    public void onEvent(ConsumerRecord<String, String> record) {
        inbound.handle(DispatchProjection.CONSUMER, record, DispatchProjection.HANDLED_EVENTS, handler::apply);
    }
}
