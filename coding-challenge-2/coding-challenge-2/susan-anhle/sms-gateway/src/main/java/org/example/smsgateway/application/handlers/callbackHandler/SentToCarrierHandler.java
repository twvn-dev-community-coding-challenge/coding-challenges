package org.example.smsgateway.application.handlers.callbackHandler;

import org.example.smsgateway.domain.model.externalevent.SentToCarrier;
import org.example.smsgateway.domain.model.message.Message;
import org.example.smsgateway.domain.model.message.MessageRepository;

public class SentToCarrierHandler implements ExternalEventHandler<SentToCarrier> {
    private final MessageRepository messageRepository;

    public SentToCarrierHandler(MessageRepository messageRepository) {
        this.messageRepository = messageRepository;
    }

    @Override
    public void handle(SentToCarrier event) {
        Message message = messageRepository.getById(event.messageId());
        message.markAsSentToCarrier();
        messageRepository.save(message);
    }
}
