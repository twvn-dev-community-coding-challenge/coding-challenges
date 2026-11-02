package org.example.smsgateway.domain.model.view;

import org.example.smsgateway.domain.model.common.DomainEvent;
import org.example.smsgateway.domain.model.common.EventRepository;
import org.example.smsgateway.domain.model.message.event.RoutePlanCalculated;
import org.example.smsgateway.domain.model.message.event.smsStatus.CarrierRejected;
import org.example.smsgateway.domain.model.message.event.smsStatus.SentSuccessfully;

import java.util.*;

public class MessageDeliveryViewProjector {

    private final EventRepository eventRepository;

    public MessageDeliveryViewProjector(EventRepository eventRepository) {
        this.eventRepository = eventRepository;
    }

    public List<DeliveryRatePerProvider> projectDeliveryRatePerProvider() {
        Map<String, MessageDeliveryView> views = project();
        List<DeliveryRatePerProvider> result = new ArrayList<>();
        for (MessageDeliveryView view : views.values()) {
            DeliveryRatePerProvider rate = findOrCreate(result, view.providerId);
            rate.incrementSent();
            if (view.succeeded) rate.incrementSucceeded();
            if (view.failed) rate.incrementFailed();
        }
        return result;
    }

    private Map<String, MessageDeliveryView> project() {
        List<DomainEvent> domainEvents = eventRepository.allEvents();
        Map<String, MessageDeliveryView> result = new HashMap<>();
        for (DomainEvent event : domainEvents) {
            switch (event) {
                case RoutePlanCalculated e -> {
                    MessageDeliveryView view = new MessageDeliveryView();
                    view.messageId = e.messageId();
                    view.providerId = e.providerId();
                    result.put(e.messageId(), view);
                }
                case SentSuccessfully e -> {
                    MessageDeliveryView view = result.get(e.messageId());
                    if (view != null) view.succeeded = true;
                }
                case CarrierRejected e -> {
                    MessageDeliveryView view = result.get(e.messageId());
                    if (view != null) view.failed = true;
                }
                default -> {}
            }
        }
        return result;
    }

    private static DeliveryRatePerProvider findOrCreate(List<DeliveryRatePerProvider> list, String providerId) {
        return list.stream()
                .filter(r -> providerId.equalsIgnoreCase(r.getProviderId()))
                .findFirst()
                .orElseGet(() -> {
                    DeliveryRatePerProvider r = new DeliveryRatePerProvider(providerId);
                    list.add(r);
                    return r;
                });
    }
}
