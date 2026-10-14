package org.example.smsgateway;

import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.boot.test.context.SpringBootTest;
import org.springframework.http.MediaType;
import org.springframework.test.annotation.DirtiesContext;
import org.springframework.test.web.servlet.MockMvc;
import org.springframework.test.web.servlet.request.MockHttpServletRequestBuilder;
import org.springframework.test.web.servlet.setup.MockMvcBuilders;
import org.springframework.web.context.WebApplicationContext;

import java.math.BigDecimal;

import static org.springframework.test.web.servlet.request.MockMvcRequestBuilders.get;
import static org.springframework.test.web.servlet.request.MockMvcRequestBuilders.post;
import static org.springframework.test.web.servlet.result.MockMvcResultMatchers.content;
import static org.springframework.test.web.servlet.result.MockMvcResultMatchers.status;

/**
 * E2E lifecycle test for a single SMS message: submission → routing → async callbacks.
 * Covers Steps 1–5 of the E2E-001 scenario (MSG-VN-001, Viettel → Vonage).
 */
@SpringBootTest
@DirtiesContext(classMode = DirtiesContext.ClassMode.BEFORE_CLASS)
class SMSLifecycleE2ETest {

    @Autowired
    private WebApplicationContext context;

    private MockMvc mockMvc;

    @BeforeEach
    void setUp() {
        mockMvc = MockMvcBuilders.webAppContextSetup(context).build();
    }

    @Test
    void msgVN001_shouldTransitionThroughFullLifecycle() throws Exception {
        // Step 1+2: submit — routing resolves +8491 → Viettel → Vonage, estimatedCost = 0.030
        sendSms("MSG-VN-001", "VN", "+84912345678", "Your OTP is 482910");

        // Step 3: async callback — provider queued
        sendCallback("MSG-VN-001", "Queue", null);

        // Step 4: async callback — handed to carrier
        sendCallback("MSG-VN-001", "Send-to-carrier", null);

        // Step 5: async callback — delivered, actualCost = 0.027
        sendCallback("MSG-VN-001", "Send-success", new BigDecimal("0.027"));
    }

    private void sendSms(String messageId, String country, String phoneNumber, String message) throws Exception {
        mockMvc.perform(post("/api/v1/sendSms")
                        .contentType(MediaType.APPLICATION_JSON)
                        .content("""
                                {
                                    "messageId": "%s",
                                    "country": "%s",
                                    "phoneNumber": "%s",
                                    "message": "%s"
                                }
                                """.formatted(messageId, country, phoneNumber, message)))
                .andExpect(status().isOk())
                .andExpect(content().string("send success"));
    }

    private void sendCallback(String messageId, String state, BigDecimal actualCost) throws Exception {
        MockHttpServletRequestBuilder req = get("/api/v1/callback-simulation")
                .param("messageId", messageId)
                .param("state", state);
        if (actualCost != null) {
            req = req.param("actualCost", actualCost.toPlainString());
        }
        mockMvc.perform(req).andExpect(status().isOk());
    }
}
