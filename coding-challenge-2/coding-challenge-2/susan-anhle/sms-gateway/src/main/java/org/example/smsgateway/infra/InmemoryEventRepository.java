package org.example.smsgateway.infra;

import java.time.Clock;
import java.time.Instant;
import java.util.ArrayList;
import java.util.Collection;
import java.util.List;
import java.util.Map;
import java.util.concurrent.ConcurrentHashMap;

import org.example.smsgateway.domain.model.common.DomainEvent;
import org.example.smsgateway.domain.model.common.EventRepository;
import tools.jackson.databind.ObjectMapper;

public class InmemoryEventRepository implements EventRepository {

  private final Clock clock;
  private final Map<String, List<SerializedEvent>> events = new ConcurrentHashMap<>();
  private final ObjectMapper jsonDeSerializer = new ObjectMapper();

    public InmemoryEventRepository(Clock clock) {
        this.clock = clock;
    }

    @Override
  public void addAll(String streamId, List<? extends DomainEvent> domainEvents) {
    List<SerializedEvent> newSerializedEvents =
        domainEvents.stream()
            .map(messageEvent -> serialize( messageEvent))
            .toList();
    events.merge(
        streamId,
        newSerializedEvents,
        (existing, added) -> {
          List<SerializedEvent> merged = new ArrayList<>(existing);
          merged.addAll(added);
          return merged;
        });
  }

  @Override
  public List<DomainEvent> get(String id) {
    List<SerializedEvent> serializedEvents = events.get(id);
    if (serializedEvents == null || serializedEvents.isEmpty()) {
      throw new IllegalArgumentException("Object not found: " + id);
    }
    List<DomainEvent> list = new ArrayList<>();
    for (SerializedEvent serializedEvent : serializedEvents) {
      DomainEvent deserialize = deserialize(serializedEvent);
      list.add(deserialize);
    }
    return list;
  }

  record SerializedEvent(String eventType, String data, Instant occuredAt) {}

  private SerializedEvent serialize(DomainEvent domainEvent) {
    String eventType = domainEvent.getClass().getName();

    String payload = jsonDeSerializer.writeValueAsString(domainEvent);
    return new SerializedEvent(eventType, payload, clock.instant());
  }



  @Override
  public List<DomainEvent> allEvents() {
    return events.values().stream()
        .flatMap(Collection::stream)
        .map(e -> this.<DomainEvent>deserialize(e))
        .toList();
  }

  private <T extends DomainEvent> T deserialize(SerializedEvent serializedEvent) {
    try {
      String eventType = serializedEvent.eventType;
      Object deserializedEvent =
          jsonDeSerializer.readValue(serializedEvent.data, Class.forName(eventType));
      return (T) deserializedEvent;
    } catch (ClassNotFoundException e) {
      throw new RuntimeException(e);
    }
  }
}
