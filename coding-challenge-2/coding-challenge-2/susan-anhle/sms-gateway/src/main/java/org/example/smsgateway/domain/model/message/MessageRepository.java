package org.example.smsgateway.domain.model.message;

import org.example.smsgateway.domain.model.common.DomainEvent;

import java.util.List;

public interface MessageRepository {
  void save(Message aggregate);

  Message getById(String id);
}
