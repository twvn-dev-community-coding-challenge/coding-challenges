package org.example.smsgateway.listener.rest;

import org.example.smsgateway.domain.model.view.DeliveryRatePerProvider;
import org.example.smsgateway.domain.model.view.MessageDeliveryViewProjector;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.RestController;

import java.util.List;

@RestController
public class GetDeliveryRatePerProvider {

    private final MessageDeliveryViewProjector messageDeliveryViewProjector;

    public GetDeliveryRatePerProvider(MessageDeliveryViewProjector messageDeliveryViewProjector) {
        this.messageDeliveryViewProjector = messageDeliveryViewProjector;
    }

    @GetMapping("/api/v1/deliveryRatePerProvider")
    public ResponseEntity<List<DeliveryRatePerProvider>> getDeliveryRatePerProvider() {
        return ResponseEntity.ok(messageDeliveryViewProjector.projectDeliveryRatePerProvider());
    }
}
