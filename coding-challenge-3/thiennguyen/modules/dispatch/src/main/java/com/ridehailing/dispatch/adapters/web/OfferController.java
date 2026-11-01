package com.ridehailing.dispatch.adapters.web;

import com.ridehailing.dispatch.adapters.web.model.OfferResponse;
import com.ridehailing.dispatch.application.OfferService;
import lombok.RequiredArgsConstructor;
import org.springframework.web.bind.annotation.*;

import java.util.List;
import java.util.UUID;

/**
 * Driver responses to offers. Dispatch owns offers, so it answers with the offer; the ride reflects the outcome
 * (DRIVER_ACCEPTED, or back to MATCHING) moments later from the published event.
 */
@RestController
@RequestMapping("/rides/{rideId}/offers")
@RequiredArgsConstructor
public class OfferController {

    private final OfferService offers;

    @PostMapping("/{driverId}/accept")
    public OfferResponse accept(@PathVariable UUID rideId, @PathVariable String driverId) {
        return OfferResponse.from(offers.accept(rideId, driverId));
    }

    @PostMapping("/{driverId}/reject")
    public OfferResponse reject(@PathVariable UUID rideId, @PathVariable String driverId) {
        return OfferResponse.from(offers.reject(rideId, driverId));
    }

    @GetMapping
    public List<OfferResponse> offers(@PathVariable UUID rideId) {
        return offers.offersFor(rideId).stream().map(OfferResponse::from).toList();
    }
}
