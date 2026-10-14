package org.example.smsgateway.application.handlers.smsHandler;

import org.example.smsgateway.application.handlers.CommandHandler;
import org.example.smsgateway.domain.model.agreement.Provider;
import org.example.smsgateway.domain.model.message.Message;
import org.example.smsgateway.domain.model.message.MessageRepository;
import org.example.smsgateway.domain.model.command.SendSms;
import org.example.smsgateway.domain.model.common.Result;
import org.example.smsgateway.domain.model.message.RoutePlan;
import org.example.smsgateway.domain.service.CarrierIdResolver;
import org.example.smsgateway.domain.service.ProviderMessageDispatcher;
import org.example.smsgateway.domain.service.routingToProvider.RouteToProviderCalculator;

public class SendSmsHandler implements CommandHandler<SendSms> {
    private final MessageRepository messageRepository;
    private final RouteToProviderCalculator routeToProviderCalculator;
    private final CarrierIdResolver carrierIdResolver;
    private final ProviderMessageDispatcher providerMessageDispatcher;

    public SendSmsHandler(MessageRepository messageRepository, RouteToProviderCalculator routeToProviderCalculator, CarrierIdResolver carrierIdResolver, ProviderMessageDispatcher providerMessageDispatcher) {
        this.messageRepository = messageRepository;
        this.routeToProviderCalculator = routeToProviderCalculator;
        this.carrierIdResolver = carrierIdResolver;
        this.providerMessageDispatcher = providerMessageDispatcher;
    }

    public Result handle(SendSms sendSms) {
        try {
            var message = new Message(sendSms.getMessageId(), sendSms.getMessage(), sendSms.getCountry(), sendSms.getPhoneNumber());
            RoutePlan routePlan = routeToProviderCalculator.calculateRoute(sendSms.getCountry(), resolveCarrierId(sendSms));
            message.addRoutePlan(routePlan);

            providerMessageDispatcher.dispatch(message, new Provider(routePlan.providerId()));
            message.markAsSentToProvider();
            messageRepository.save(message);
            return Result.success();
        } catch (Exception e) {
            return Result.failure(e);
        }
    }

    private String resolveCarrierId(SendSms sendSms) {
        return carrierIdResolver.determineCarrierId(sendSms.getPhoneNumber());
    }
}
