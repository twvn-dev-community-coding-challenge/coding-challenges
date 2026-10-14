package org.example.smsgateway.infra;

import org.example.smsgateway.domain.model.message.Message;
import org.example.smsgateway.domain.service.ProviderMessageDispatcher;

public class TwilioSendMessageGateway implements ProviderMessageDispatcher.ProviderSendMessageGateway {
    @Override
    public void handle(Message message) {
        // send to azure service bus queue: connection String: servicebus.azure.com/twilio-send;secretId:sercretId;secretName:secretName:secretName
    }
}
