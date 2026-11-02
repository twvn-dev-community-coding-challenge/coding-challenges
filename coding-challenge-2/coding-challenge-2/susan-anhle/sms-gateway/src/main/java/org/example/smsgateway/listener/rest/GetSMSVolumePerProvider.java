package org.example.smsgateway.listener.rest;

import org.example.smsgateway.domain.model.view.MessageVolumeViewProjector;
import org.example.smsgateway.domain.model.view.SMSVolumePerProvider;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.RestController;

import java.util.List;

@RestController
public class GetSMSVolumePerProvider {

    private final MessageVolumeViewProjector messageVolumeViewProjector;

    public GetSMSVolumePerProvider(MessageVolumeViewProjector messageVolumeViewProjector) {
        this.messageVolumeViewProjector = messageVolumeViewProjector;
    }

    @GetMapping("/api/v1/smsVolumePerProvider")
    public ResponseEntity<List<SMSVolumePerProvider>> getSMSVolumePerProvider() {
        return ResponseEntity.ok(messageVolumeViewProjector.projectSMSVolumePerProvider());
    }
}
