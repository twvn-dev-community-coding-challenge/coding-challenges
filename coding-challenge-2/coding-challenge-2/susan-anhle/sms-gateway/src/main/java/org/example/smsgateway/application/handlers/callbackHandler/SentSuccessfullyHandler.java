package org.example.smsgateway.application.handlers.callbackHandler;

import org.example.smsgateway.domain.model.message.Message;
import org.example.smsgateway.domain.model.message.MessageRepository;
import org.example.smsgateway.domain.model.externalevent.SentSuccessfully;

public class SentSuccessfullyHandler implements ExternalEventHandler<SentSuccessfully> {
    private final MessageRepository messageRepository;

    public SentSuccessfullyHandler(MessageRepository messageRepository) {
        this.messageRepository = messageRepository;
    }

    @Override
    public void handle(SentSuccessfully event) {
        Message message = messageRepository.getById(event.messageId());
        message.markAsSentSuccessfully(event.actualCost());
        messageRepository.save(message);
    }
}
