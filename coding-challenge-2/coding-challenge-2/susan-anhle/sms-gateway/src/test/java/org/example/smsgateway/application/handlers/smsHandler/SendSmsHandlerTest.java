package org.example.smsgateway.application.handlers.smsHandler;

import org.example.smsgateway.application.handlers.smsHandler.SendSmsHandler;
import org.example.smsgateway.domain.model.agreement.Provider;
import org.example.smsgateway.domain.model.command.SendSms;
import org.example.smsgateway.domain.model.common.Result;
import org.example.smsgateway.domain.model.message.Message;
import org.example.smsgateway.domain.model.message.MessageRepository;
import org.example.smsgateway.domain.model.message.RoutePlan;
import org.example.smsgateway.domain.service.CarrierIdResolver;
import org.example.smsgateway.domain.service.ProviderMessageDispatcher;
import org.example.smsgateway.domain.service.routingToProvider.RouteToProviderCalculator;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.ArgumentCaptor;
import org.mockito.InOrder;
import org.mockito.Mock;
import org.mockito.junit.jupiter.MockitoExtension;

import java.math.BigDecimal;

import static org.assertj.core.api.Assertions.assertThat;
import static org.mockito.ArgumentMatchers.any;
import static org.mockito.Mockito.*;

@ExtendWith(MockitoExtension.class)
class SendSmsHandlerTest {

    @Mock
    private MessageRepository messageRepository;

    @Mock
    private RouteToProviderCalculator routeToProviderCalculator;

    @Mock
    private CarrierIdResolver carrierIdResolver;

    @Mock
    private ProviderMessageDispatcher providerMessageDispatcher;

    private SendSmsHandler handler;

    private static final String MESSAGE_ID = "msg-001";
    private static final String MESSAGE_TEXT = "This is a test message";
    private static final String COUNTRY = "VN";
    private static final String PHONE_NUMBER = "+12025551234";
    private static final String CARRIER_ID = "carrier-test";
    private static final String PROVIDER_ID = "provider-test";
    private static final BigDecimal ESTIMATED_COST = new BigDecimal("0.05");

    private void mockedCarrierResolver() {
        when(carrierIdResolver.determineCarrierId(PHONE_NUMBER)).thenReturn(CARRIER_ID);
    }

    private void mockedRouteToProviderCalculator() {
        when(routeToProviderCalculator.calculateRoute(COUNTRY, CARRIER_ID))
                .thenReturn(new RoutePlan(PROVIDER_ID, ESTIMATED_COST));
    }

    private SendSms validSendSmsCommand() {
        return new SendSms(MESSAGE_ID, MESSAGE_TEXT, COUNTRY, PHONE_NUMBER);
    }

    @BeforeEach
    void setUp() {
        handler = new SendSmsHandler(messageRepository, routeToProviderCalculator, carrierIdResolver, providerMessageDispatcher);
    }

    @Test
    void shouldReturnSuccess() {
        mockedCarrierResolver();
        mockedRouteToProviderCalculator();

        Result result = handler.handle(validSendSmsCommand());

        assertThat(result.isSuccess()).isTrue();
    }

    @Test
    void shouldDispatchMessageToCorrectProvider() {
        mockedCarrierResolver();
        mockedRouteToProviderCalculator();

        handler.handle(validSendSmsCommand());

        ArgumentCaptor<Provider> providerCaptor = ArgumentCaptor.forClass(Provider.class);
        verify(providerMessageDispatcher).dispatch(any(Message.class), providerCaptor.capture());
        assertThat(providerCaptor.getValue().getId()).isEqualTo(PROVIDER_ID);
    }

    @Test
    void shouldSaveMessageWithCorrectId() {
        mockedCarrierResolver();
        mockedRouteToProviderCalculator();

        handler.handle(validSendSmsCommand());

        ArgumentCaptor<Message> messageCaptor = ArgumentCaptor.forClass(Message.class);
        verify(messageRepository).save(messageCaptor.capture());
        assertThat(messageCaptor.getValue().id()).isEqualTo(MESSAGE_ID);
    }

    @Test
    void shouldNotSaveMessageAndNotDispatchToProviderWhenFailedToCalculateRoute() {
        mockedCarrierResolver();
        when(routeToProviderCalculator.calculateRoute(COUNTRY, CARRIER_ID)).thenThrow(new RuntimeException("Failed to calculate route"));

        Result result = handler.handle(validSendSmsCommand());

        assertThat(result.isSuccess()).isFalse();
        verify(messageRepository, never()).save(any(Message.class));
        verify(providerMessageDispatcher, never()).dispatch(any(Message.class), any(Provider.class));
    }
}
