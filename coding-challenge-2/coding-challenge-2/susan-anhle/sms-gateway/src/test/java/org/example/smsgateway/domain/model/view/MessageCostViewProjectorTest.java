package org.example.smsgateway.domain.model.view;

import org.example.smsgateway.domain.model.message.event.NewMessageRequestReceived;
import org.example.smsgateway.domain.model.message.event.RoutePlanCalculated;
import org.example.smsgateway.domain.model.message.event.smsStatus.SentSuccessfully;
import org.example.smsgateway.infra.InmemoryEventRepository;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;

import java.math.BigDecimal;
import java.util.List;

import static org.assertj.core.api.Assertions.assertThat;
import static org.mockito.Mockito.mock;
import static org.mockito.Mockito.when;

class MessageCostViewProjectorTest {

  private MessageCostViewProjector projector;
  private InmemoryEventRepository events;

  @BeforeEach
  void setUp() {
    events = mock(InmemoryEventRepository.class);
    projector = new MessageCostViewProjector(events);
  }

  @Test
  void projectCostPerCountry_shouldReturnEmptyList_whenNoEvents() {
    List<CostPerCountry> result = projector.projectCostPerCountry();

    assertThat(result).isEmpty();
  }

  @Test
  void projectCostPerCountry_shouldReturnOneEntry_whenSingleMessageWithEstimatedCostOnly() {
    when(events.allEvents())
        .thenReturn(
            List.of(
                new NewMessageRequestReceived("msg-1", "Hello", "VN", "+84901234567"),
                new RoutePlanCalculated("msg-1", "provider-twilio", new BigDecimal("0.05"))));

    List<CostPerCountry> result = projector.projectCostPerCountry();

    assertThat(result)
        .containsExactlyInAnyOrder(new CostPerCountry("VN", new BigDecimal("0.05"), null));
  }

  @Test
  void projectCostPerCountry_shouldIncludeActualCost_whenSentSuccessfullyEventPresent() {

    when(events.allEvents())
        .thenReturn(
            List.of(
                new NewMessageRequestReceived("msg-1", "Hello", "VN", "+84901234567"),
                new RoutePlanCalculated("msg-1", "provider-twilio", new BigDecimal("0.05")),
                new SentSuccessfully("msg-1", new BigDecimal("0.04"))));
    List<CostPerCountry> result = projector.projectCostPerCountry();

    assertThat(result)
        .containsExactlyInAnyOrder(
            new CostPerCountry("VN", new BigDecimal("0.05"), new BigDecimal("0.04")));
  }

  @Test
  void projectCostPerCountry_shouldAggregateEstimatedCosts_whenMultipleMessagesFromSameCountry() {
    when(events.allEvents())
        .thenReturn(
            List.of(
                new NewMessageRequestReceived("msg-1", "Hello", "VN", "+84901234567"),
                new RoutePlanCalculated("msg-1", "provider-twilio", new BigDecimal("0.05")),
                new SentSuccessfully("msg-1", new BigDecimal("0.04")),
                new NewMessageRequestReceived("msg-2", "World", "VN", "+84907654321"),
                new RoutePlanCalculated("msg-2", "provider-twilio", new BigDecimal("0.10")),
                new SentSuccessfully("msg-2", new BigDecimal("0.09"))));
    List<CostPerCountry> result = projector.projectCostPerCountry();

    assertThat(result)
        .containsExactlyInAnyOrder(
            new CostPerCountry("VN", new BigDecimal("0.15"), new BigDecimal("0.13")));
  }

  @Test
  void projectCostPerCountry_shouldReturnSeparateEntries_whenMessagesFromDifferentCountries() {
    when(events.allEvents())
        .thenReturn(List.of(
            new NewMessageRequestReceived("msg-1", "Hello", "VN", "+84901234567"),
            new RoutePlanCalculated("msg-1", "provider-twilio", new BigDecimal("0.05")),
            new NewMessageRequestReceived("msg-2", "World", "US", "+12025551234"),
            new RoutePlanCalculated("msg-2", "provider-vonage", new BigDecimal("0.08")),
            new SentSuccessfully("msg-1", new BigDecimal("0.04")),
            new NewMessageRequestReceived("msg-3", "OTP", "VN", "+84901234567"),
            new RoutePlanCalculated("msg-3", "provider-obiwan", new BigDecimal("0.18")),
            new SentSuccessfully("msg-2", new BigDecimal("0.07"))));
    List<CostPerCountry> result = projector.projectCostPerCountry();

    assertThat(result)
        .containsExactlyInAnyOrder(
            new CostPerCountry("VN", new BigDecimal("0.23"), new BigDecimal("0.04")),
            new CostPerCountry("US", new BigDecimal("0.08"), new BigDecimal("0.07")));
  }
}
