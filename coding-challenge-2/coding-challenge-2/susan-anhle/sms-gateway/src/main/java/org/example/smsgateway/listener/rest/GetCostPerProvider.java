package org.example.smsgateway.listener.rest;

import org.example.smsgateway.domain.model.common.DomainEvent;
import org.example.smsgateway.domain.model.message.MessageRepository;
import org.example.smsgateway.domain.model.view.CostPerProvider;
import org.example.smsgateway.domain.model.view.MessageCostViewProjector;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.RestController;

import java.util.*;

@RestController
public class GetCostPerProvider {

    private final MessageCostViewProjector messageCostViewProjector;

    public GetCostPerProvider(MessageCostViewProjector messageCostViewProjector) {
        this.messageCostViewProjector = messageCostViewProjector;
    }

    @GetMapping("/api/v1/costPerProvider")
    public ResponseEntity<List<CostPerProvider>> getCostPerProvider() {
        return ResponseEntity.ok(messageCostViewProjector.projectCostPerProviderId());
    }
}
