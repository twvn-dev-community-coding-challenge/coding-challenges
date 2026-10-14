package org.example.smsgateway;

import org.example.smsgateway.application.handlers.smsHandler.SendSmsHandler;
import org.example.smsgateway.domain.model.command.SendSms;
import org.example.smsgateway.domain.model.externalevent.CarrierRejected;
import org.example.smsgateway.domain.model.externalevent.Queued;
import org.example.smsgateway.domain.model.externalevent.SentSuccessfully;
import org.example.smsgateway.domain.model.externalevent.SentToCarrier;
import org.example.smsgateway.domain.service.ExternalEventHandlerDispatcher;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.boot.test.context.SpringBootTest;
import org.springframework.test.annotation.DirtiesContext;
import org.springframework.test.web.servlet.MockMvc;
import org.springframework.test.web.servlet.setup.MockMvcBuilders;
import org.springframework.web.context.WebApplicationContext;

import java.math.BigDecimal;

import static org.hamcrest.Matchers.hasItem;
import static org.springframework.test.web.servlet.request.MockMvcRequestBuilders.get;
import static org.springframework.test.web.servlet.result.MockMvcResultMatchers.jsonPath;
import static org.springframework.test.web.servlet.result.MockMvcResultMatchers.status;

/**
 * E2E report tests — Steps 6–9 of the E2E-001 scenario.
 * Scenario: 4 VN messages (1 Viettel/Vonage + 3 Vinaphone/Twilio), all delivered.
 * Step 9 adds MSG-VN-005 (Vonage, carrier-rejected) to test delivery rate.
 * Each test gets a fresh context so reports reflect only the scenario's messages.
 */
@SpringBootTest
@DirtiesContext(classMode = DirtiesContext.ClassMode.BEFORE_EACH_TEST_METHOD)
class SMSReportsE2ETest {

    @Autowired
    private WebApplicationContext context;

    @Autowired
    private SendSmsHandler sendSmsHandler;

    @Autowired
    private ExternalEventHandlerDispatcher externalEventHandlerDispatcher;

    private MockMvc mockMvc;

    @BeforeEach
    void setUp() {
        mockMvc = MockMvcBuilders.webAppContextSetup(context).build();

        // MSG-VN-001: Viettel → Vonage, estimatedCost = 0.030, actualCost = 0.027
        deliver("MSG-VN-001", "VN", "+84912345678", "Your OTP is 482910", new BigDecimal("0.027"));

        // MSG-VN-002/003/004: Vinaphone → Twilio, estimatedCost = 0.015 each
        deliver("MSG-VN-002", "VN", "+84882345670", "test", new BigDecimal("0.013"));
        deliver("MSG-VN-003", "VN", "+84882345671", "test", new BigDecimal("0.013"));
        deliver("MSG-VN-004", "VN", "+84882345672", "test", new BigDecimal("0.014"));
    }

    // Step 6
    @Test
    void costPerCountry_shouldAggregateCostByCountry() throws Exception {
        mockMvc.perform(get("/api/v1/costPerCountry"))
                .andExpect(status().isOk())
                .andExpect(jsonPath("$[0].country").value("VN"))
                .andExpect(jsonPath("$[0].estimatedCost").value(0.075))  // 0.030 + 3×0.015
                .andExpect(jsonPath("$[0].actualCost").value(0.067));    // 0.027 + 0.013 + 0.013 + 0.014
    }

    // Step 7
    @Test
    void costPerProvider_shouldAggregateCostByProvider() throws Exception {
        mockMvc.perform(get("/api/v1/costPerProvider"))
                .andExpect(status().isOk())
                .andExpect(jsonPath("$[?(@.providerId == 'Vonage')].estimatedCost", hasItem(0.03)))
                .andExpect(jsonPath("$[?(@.providerId == 'Vonage')].actualCost", hasItem(0.027)))
                .andExpect(jsonPath("$[?(@.providerId == 'Twilio')].estimatedCost", hasItem(0.045)))
                .andExpect(jsonPath("$[?(@.providerId == 'Twilio')].actualCost", hasItem(0.04)));
    }

    // Step 8
    @Test
    void smsVolumePerProvider_shouldCountRoutedMessagesByProvider() throws Exception {
        mockMvc.perform(get("/api/v1/smsVolumePerProvider"))
                .andExpect(status().isOk())
                .andExpect(jsonPath("$[?(@.providerId == 'Vonage')].messageCount", hasItem(1)))
                .andExpect(jsonPath("$[?(@.providerId == 'Twilio')].messageCount", hasItem(3)));
    }

    // Step 9
    @Test
    void deliveryRatePerProvider_shouldReportSuccessAndFailureRates() throws Exception {
        // MSG-VN-005: Viettel → Vonage, carrier-rejected
        reject("MSG-VN-005", "VN", "+84912345679", "test");

        // Vonage: 2 sent (MSG-VN-001 succeeded, MSG-VN-005 rejected) → successRate=0.5, failureRate=0.5
        // Twilio: 3 sent (all succeeded) → successRate=1.0, failureRate=0.0
        mockMvc.perform(get("/api/v1/deliveryRatePerProvider"))
                .andExpect(status().isOk())
                .andExpect(jsonPath("$[?(@.providerId == 'Vonage')].sent", hasItem(2)))
                .andExpect(jsonPath("$[?(@.providerId == 'Vonage')].succeeded", hasItem(1)))
                .andExpect(jsonPath("$[?(@.providerId == 'Vonage')].failed", hasItem(1)))
                .andExpect(jsonPath("$[?(@.providerId == 'Vonage')].successRate", hasItem(0.5)))
                .andExpect(jsonPath("$[?(@.providerId == 'Vonage')].failureRate", hasItem(0.5)))
                .andExpect(jsonPath("$[?(@.providerId == 'Twilio')].sent", hasItem(3)))
                .andExpect(jsonPath("$[?(@.providerId == 'Twilio')].succeeded", hasItem(3)))
                .andExpect(jsonPath("$[?(@.providerId == 'Twilio')].failed", hasItem(0)))
                .andExpect(jsonPath("$[?(@.providerId == 'Twilio')].successRate", hasItem(1.0)))
                .andExpect(jsonPath("$[?(@.providerId == 'Twilio')].failureRate", hasItem(0.0)));
    }

    private void deliver(String messageId, String country, String phoneNumber, String message, BigDecimal actualCost) {
        sendSmsHandler.handle(new SendSms(messageId, message, country, phoneNumber));
        externalEventHandlerDispatcher.dispatch(new Queued(messageId));
        externalEventHandlerDispatcher.dispatch(new SentToCarrier(messageId));
        externalEventHandlerDispatcher.dispatch(new SentSuccessfully(messageId, actualCost));
    }

    private void reject(String messageId, String country, String phoneNumber, String message) {
        sendSmsHandler.handle(new SendSms(messageId, message, country, phoneNumber));
        externalEventHandlerDispatcher.dispatch(new Queued(messageId));
        externalEventHandlerDispatcher.dispatch(new CarrierRejected(messageId));
    }
}
