package org.example.smsgateway.listener.rest;

import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.boot.test.context.SpringBootTest;
import org.springframework.http.MediaType;
import org.springframework.test.web.servlet.MockMvc;
import org.springframework.test.web.servlet.setup.MockMvcBuilders;
import org.springframework.web.context.WebApplicationContext;

import static org.springframework.test.web.servlet.request.MockMvcRequestBuilders.post;
import static org.springframework.test.web.servlet.result.MockMvcResultMatchers.content;
import static org.springframework.test.web.servlet.result.MockMvcResultMatchers.status;

@SpringBootTest
class PostSMSIntegrationTest {

    @Autowired
    private WebApplicationContext context;

    private MockMvc mockMvc;

    @BeforeEach
    void setUp() {
        mockMvc = MockMvcBuilders.webAppContextSetup(context).build();
    }

    // +8491 prefix → Viettel carrier → VN country → Vonage provider (registered in ProviderMessageDispatcher)
    @Test
    void sendSms_shouldReturn200AndSendSuccess_whenRequestIsValid() throws Exception {
        mockMvc.perform(post("/api/v1/sendSms")
                        .contentType(MediaType.APPLICATION_JSON)
                        .content("""
                                {
                                    "messageId": "integration-msg-001",
                                    "message": "Hello World",
                                    "country": "VN",
                                    "phoneNumber": "+84912345678"
                                }
                                """))
                .andExpect(status().isOk())
                .andExpect(content().string("send success"));
    }

    // +8488 prefix → Vinaphone carrier → VN country → Twilio provider (registered in ProviderMessageDispatcher)
    @Test
    void sendSms_shouldReturn200_whenRoutesToTwilio() throws Exception {
        mockMvc.perform(post("/api/v1/sendSms")
                        .contentType(MediaType.APPLICATION_JSON)
                        .content("""
                                {
                                    "messageId": "integration-msg-002",
                                    "message": "Hello World",
                                    "country": "VN",
                                    "phoneNumber": "+84882345678"
                                }
                                """))
                .andExpect(status().isOk())
                .andExpect(content().string("send success"));
    }

    @Test
    void sendSms_shouldReturn500_whenPhoneNumberPrefixHasNoKnownCarrier() throws Exception {
        mockMvc.perform(post("/api/v1/sendSms")
                        .contentType(MediaType.APPLICATION_JSON)
                        .content("""
                                {
                                    "messageId": "integration-msg-003",
                                    "message": "Hello World",
                                    "country": "US",
                                    "phoneNumber": "+12025550000"
                                }
                                """))
                .andExpect(status().isInternalServerError());
    }

    @Test
    void sendSms_shouldReturn500_whenNoAgreementExistsForCarrierAndCountry() throws Exception {
        // +8491 resolves to Viettel, but "TH" has no agreement for Viettel
        mockMvc.perform(post("/api/v1/sendSms")
                        .contentType(MediaType.APPLICATION_JSON)
                        .content("""
                                {
                                    "messageId": "integration-msg-004",
                                    "message": "Hello World",
                                    "country": "TH",
                                    "phoneNumber": "+84912345678"
                                }
                                """))
                .andExpect(status().isInternalServerError());
    }

    @Test
    void sendSms_shouldReturn500_whenMessageExceeds999Characters() throws Exception {
        String longMessage = "A".repeat(1000);
        mockMvc.perform(post("/api/v1/sendSms")
                        .contentType(MediaType.APPLICATION_JSON)
                        .content("""
                                {
                                    "messageId": "integration-msg-005",
                                    "message": "%s",
                                    "country": "VN",
                                    "phoneNumber": "+84912345678"
                                }
                                """.formatted(longMessage)))
                .andExpect(status().isInternalServerError());
    }

    @Test
    void sendSms_shouldReturn500_whenRequiredFieldIsMissing() throws Exception {
        mockMvc.perform(post("/api/v1/sendSms")
                        .contentType(MediaType.APPLICATION_JSON)
                        .content("""
                                {
                                    "messageId": "integration-msg-006",
                                    "country": "VN",
                                    "phoneNumber": "+84912345678"
                                }
                                """))
                .andExpect(status().isInternalServerError());
    }
}
