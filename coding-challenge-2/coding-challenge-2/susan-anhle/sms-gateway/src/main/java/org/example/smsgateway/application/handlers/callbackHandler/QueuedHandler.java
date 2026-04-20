package org.example.smsgateway.application.handlers.callbackHandler;

import org.example.smsgateway.domain.model.externalevent.Queued;
import org.example.smsgateway.domain.model.message.Message;
import org.example.smsgateway.domain.model.message.MessageRepository;

public class QueuedHandler implements ExternalEventHandler<Queued> {
    private final MessageRepository messageRepository;

    public QueuedHandler(MessageRepository messageRepository) {
        this.messageRepository = messageRepository;
    }


    @Override
    public void handle(Queued event) {
        Message message = messageRepository.getById(event.messageId());
        message.markAsQueued();
        messageRepository.save(message);
    }
}
