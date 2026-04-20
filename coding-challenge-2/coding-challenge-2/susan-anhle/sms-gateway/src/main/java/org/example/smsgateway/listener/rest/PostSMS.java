package org.example.smsgateway.listener.rest;

import org.example.smsgateway.application.handlers.smsHandler.SendSmsHandler;
import org.example.smsgateway.domain.model.command.SendSms;
import org.example.smsgateway.domain.model.common.Result;
import org.jspecify.annotations.NonNull;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.PostMapping;
import org.springframework.web.bind.annotation.RequestBody;
import org.springframework.web.bind.annotation.RestController;

@RestController
public class PostSMS {
    private final SendSmsHandler sendSmsHandler;

    public PostSMS(SendSmsHandler sendSmsHandler) {
        this.sendSmsHandler = sendSmsHandler;
    }

    @PostMapping("/api/v1/sendSms")
    public ResponseEntity<String> postMessage(@RequestBody SmsRequest smsRequest) {
        try {
            Result handle = sendSmsHandler.handle(toCommand(smsRequest));
            if (handle.isSuccess()) return ResponseEntity.ok().body("send success");
            return ResponseEntity.internalServerError().build();
        } catch (Exception e) {
            return ResponseEntity.internalServerError().build();
        }
    }

    private static @NonNull SendSms toCommand(SmsRequest smsRequest) {
        return new SendSms(
                smsRequest.messageId, smsRequest.message,
                smsRequest.country, smsRequest.phoneNumber
        );
    }

    public record SmsRequest(
            String messageId, String country, String phoneNumber, String message
    ) {
    }
}
