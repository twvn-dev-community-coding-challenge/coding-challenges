package org.example.smsgateway.listener.rest;

import org.example.smsgateway.domain.model.common.DomainEvent;
import org.example.smsgateway.domain.model.message.MessageRepository;
import org.example.smsgateway.domain.model.view.CostPerCountry;
import org.example.smsgateway.domain.model.view.MessageCostViewProjector;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.RestController;

import java.util.*;

@RestController
public class GetCostPerCountry {

    private final MessageCostViewProjector messageCostViewProjector;

    public GetCostPerCountry(MessageCostViewProjector messageCostViewProjector) {
        this.messageCostViewProjector = messageCostViewProjector;
    }


    @GetMapping("/api/v1/costPerCountry")
    public ResponseEntity<List<CostPerCountry>> getCostPerCountry() {
        try {
            return ResponseEntity.ok(messageCostViewProjector.projectCostPerCountry());
        } catch (Exception e) {
            return ResponseEntity.internalServerError().build();
        }
    }


}
