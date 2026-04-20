package org.example.smsgateway.infra;

import org.example.smsgateway.domain.model.common.EventRepository;
import org.example.smsgateway.domain.model.message.Message;
import org.example.smsgateway.domain.model.message.MessageRepository;
import org.example.smsgateway.domain.model.message.event.MessageEvent;

import java.util.List;

public class InmemoryMessageRepository implements MessageRepository {
  private final EventRepository eventRepository;

  public InmemoryMessageRepository(EventRepository eventRepository) {
    this.eventRepository = eventRepository;
  }

  @Override
  public void save(Message message) {
    eventRepository.addAll(message.id(), message.getChanges());
    message.clearChanges();
  }

  @Override
  public Message getById(String id) {
    return new Message((List<MessageEvent>) (List<?>) eventRepository.get(id));
  }
}
