package org.example.smsgateway.application.handlers.callbackHandler;

import org.example.smsgateway.domain.model.message.Message;
import org.example.smsgateway.domain.model.message.MessageRepository;
import org.example.smsgateway.domain.model.externalevent.CarrierRejected;

public class CarrierRejectedHandler implements ExternalEventHandler<CarrierRejected> {
    private final MessageRepository messageRepository;

    public CarrierRejectedHandler(MessageRepository messageRepository) {
        this.messageRepository = messageRepository;
    }

    @Override
    public void handle(CarrierRejected event) {
        Message message = messageRepository.getById(event.messageId());
        message.markAsCarrierRejected();
        messageRepository.save(message);
    }
}
