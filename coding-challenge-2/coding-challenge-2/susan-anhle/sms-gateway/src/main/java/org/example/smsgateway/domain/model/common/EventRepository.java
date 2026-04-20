package org.example.smsgateway.domain.model.common;

import java.util.List;

public interface EventRepository {
    void addAll(String streamId, List<? extends DomainEvent> domainEvents);

    List<DomainEvent> get(String id);

    List<DomainEvent> allEvents();
}
