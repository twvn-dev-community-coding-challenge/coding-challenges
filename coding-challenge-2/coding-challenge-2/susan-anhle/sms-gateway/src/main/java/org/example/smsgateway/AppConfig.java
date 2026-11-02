package org.example.smsgateway;

import org.example.smsgateway.application.handlers.callbackHandler.CarrierRejectedHandler;
import org.example.smsgateway.application.handlers.callbackHandler.QueuedHandler;
import org.example.smsgateway.application.handlers.smsHandler.SendSmsHandler;
import org.example.smsgateway.application.handlers.callbackHandler.SentSuccessfullyHandler;
import org.example.smsgateway.application.handlers.callbackHandler.SentToCarrierHandler;
import org.example.smsgateway.domain.model.agreement.Provider;
import org.example.smsgateway.domain.model.common.EventRepository;
import org.example.smsgateway.domain.model.externalevent.CarrierRejected;
import org.example.smsgateway.domain.model.externalevent.Queued;
import org.example.smsgateway.domain.model.externalevent.SentSuccessfully;
import org.example.smsgateway.domain.model.externalevent.SentToCarrier;
import org.example.smsgateway.domain.model.message.MessageRepository;
import org.example.smsgateway.domain.service.CarrierIdResolver;
import org.example.smsgateway.domain.service.ExternalEventHandlerDispatcher;
import org.example.smsgateway.domain.service.ProviderMessageDispatcher;
import org.example.smsgateway.domain.model.view.MessageCostViewProjector;
import org.example.smsgateway.domain.model.view.MessageDeliveryViewProjector;
import org.example.smsgateway.domain.model.view.MessageVolumeViewProjector;
import org.example.smsgateway.domain.service.routingToProvider.RouteToProviderCalculator;
import org.example.smsgateway.domain.model.agreement.AgreementRepository;
import org.example.smsgateway.domain.service.CarrierIdResolverImpl;
import org.example.smsgateway.domain.service.routingToProvider.routingstrategy.CheapestProviderRouting;
import org.example.smsgateway.infra.InmemoryEventRepository;
import org.example.smsgateway.infra.InmemoryAgreementRepository;
import org.example.smsgateway.infra.InmemoryMessageRepository;
import org.example.smsgateway.infra.TwilioSendMessageGateway;
import org.example.smsgateway.infra.VonageSendMessageGateway;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;
import tools.jackson.databind.JsonNode;
import tools.jackson.databind.ObjectMapper;

import java.io.InputStream;
import java.time.Clock;
import java.util.HashMap;
import java.util.Map;

@Configuration
public class AppConfig {

  @Bean
  public EventRepository eventRepository() {
    return new InmemoryEventRepository(Clock.systemUTC());
  }

  @Bean
  public MessageRepository messageRepository(EventRepository eventRepository) {
    return new InmemoryMessageRepository(eventRepository);
  }

  @Bean
  public ExternalEventHandlerDispatcher externalEventHandlerDispatcher(
      MessageRepository messageRepository) {
    return new ExternalEventHandlerDispatcher(
        Map.of(
            CarrierRejected.class, new CarrierRejectedHandler(messageRepository),
            Queued.class, new QueuedHandler(messageRepository),
            SentSuccessfully.class, new SentSuccessfullyHandler(messageRepository),
            SentToCarrier.class, new SentToCarrierHandler(messageRepository)));
  }

  @Bean
  public ProviderMessageDispatcher setupMessageToProviderDispatcher() {
    return new ProviderMessageDispatcher(
        Map.of(
            new Provider("Twilio"), new TwilioSendMessageGateway(),
            new Provider("Vonage"), new VonageSendMessageGateway()));
  }

  @Bean
  public MessageCostViewProjector messageCostViewProjector(EventRepository eventRepository) {
    return new MessageCostViewProjector(eventRepository);
  }

  @Bean
  public MessageVolumeViewProjector messageVolumeViewProjector(EventRepository eventRepository) {
    return new MessageVolumeViewProjector(eventRepository);
  }

  @Bean
  public MessageDeliveryViewProjector messageDeliveryViewProjector(EventRepository eventRepository) {
    return new MessageDeliveryViewProjector(eventRepository);
  }

  @Bean
  public CarrierIdResolver carrierIdResolver() {
    try (InputStream is = getClass().getResourceAsStream("/phone-carrier-prefixes.json")) {
      JsonNode nodes = new ObjectMapper().readTree(is);
      Map<String, String> prefixToCarrierId = new HashMap<>();
      for (JsonNode node : nodes) {
        prefixToCarrierId.put(node.get("prefix").asText(), node.get("carrierId").asText());
      }
      return new CarrierIdResolverImpl(prefixToCarrierId);
    } catch (Exception e) {
      throw new IllegalStateException("Failed to load phone-carrier-prefixes.json", e);
    }
  }

  @Bean
  public AgreementRepository agreementRepository() {
    return new InmemoryAgreementRepository();
  }

  @Bean
  public RouteToProviderCalculator routeToProviderCalculator(
      AgreementRepository agreementRepository) {
    return new CheapestProviderRouting(agreementRepository);
  }

  @Bean
  public SendSmsHandler sendSmsHandler(
      MessageRepository messageRepository,
      RouteToProviderCalculator routeToProviderCalculator,
      CarrierIdResolver carrierIdResolver,
      ProviderMessageDispatcher providerMessageDispatcher) {
    return new SendSmsHandler(
        messageRepository, routeToProviderCalculator, carrierIdResolver, providerMessageDispatcher);
  }
}
