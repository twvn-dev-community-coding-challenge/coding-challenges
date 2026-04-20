package org.example.smsgateway.domain.model.view;

import org.example.smsgateway.domain.model.common.DomainEvent;
import org.example.smsgateway.domain.model.common.EventRepository;
import org.example.smsgateway.domain.model.message.event.RoutePlanCalculated;

import java.util.*;

public class MessageVolumeViewProjector {

    private final EventRepository eventRepository;

    public MessageVolumeViewProjector(EventRepository eventRepository) {
        this.eventRepository = eventRepository;
    }

    public List<SMSVolumePerProvider> projectSMSVolumePerProvider() {
        Map<String, MessageVolumeView> views = project();
        List<SMSVolumePerProvider> result = new ArrayList<>();
        for (MessageVolumeView view : views.values()) {
            if (view.providerId == null) continue;
            Optional<SMSVolumePerProvider> existing = result.stream()
                    .filter(v -> view.providerId.equalsIgnoreCase(v.getProviderId()))
                    .findFirst();
            if (existing.isEmpty()) {
                result.add(new SMSVolumePerProvider(view.providerId, 1));
            } else {
                existing.get().incrementMessageCount();
            }
        }
        return result;
    }

    private Map<String, MessageVolumeView> project() {
        List<DomainEvent> domainEvents = eventRepository.allEvents();
        Map<String, MessageVolumeView> result = new HashMap<>();
        for (DomainEvent event : domainEvents) {
            if (event instanceof RoutePlanCalculated e) {
                MessageVolumeView view = new MessageVolumeView();
                view.messageId = e.messageId();
                view.providerId = e.providerId();
                result.put(e.messageId(), view);
            }
        }
        return result;
    }
}
