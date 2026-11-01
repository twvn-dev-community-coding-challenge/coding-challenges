package com.ridehailing.ride.adapters.web;

import com.ridehailing.common.error.ValidationException;
import com.ridehailing.ride.adapters.web.model.CreateRideRequest;
import com.ridehailing.ride.adapters.web.model.RideResponse;
import com.ridehailing.ride.application.RideCommandService;
import com.ridehailing.ride.application.RideCreationService;
import com.ridehailing.ride.application.RideLifecycle;
import com.ridehailing.ride.domain.RideHistoryEntry;
import jakarta.validation.Valid;
import lombok.RequiredArgsConstructor;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

import java.net.URI;
import java.util.List;
import java.util.UUID;

@RestController
@RequestMapping("/rides")
@RequiredArgsConstructor
public class RideController {

    public static final String IDEMPOTENCY_KEY = "Idempotency-Key";
    public static final String IDEMPOTENT_REPLAYED = "Idempotent-Replayed";
    private static final int MAX_IDEMPOTENCY_KEY_LENGTH = 128;

    private final RideCreationService creation;
    private final RideCommandService commands;
    private final RideLifecycle rides;

    @PostMapping
    public ResponseEntity<RideResponse> create(
            @RequestHeader(name = IDEMPOTENCY_KEY, required = false) String idempotencyKey,
            @Valid @RequestBody CreateRideRequest request) {
        if (idempotencyKey == null || idempotencyKey.isBlank()) {
            throw new ValidationException("IDEMPOTENCY_KEY_REQUIRED", "The Idempotency-Key header is required");
        }
        if (idempotencyKey.length() > MAX_IDEMPOTENCY_KEY_LENGTH) {
            throw new ValidationException("IDEMPOTENCY_KEY_TOO_LONG",
                    "The Idempotency-Key header must be at most %d characters".formatted(MAX_IDEMPOTENCY_KEY_LENGTH));
        }
        RideCreationService.CreationResult result = creation.create(idempotencyKey, request.toCommand());
        return ResponseEntity.created(URI.create("/rides/" + result.ride().id()))
                .header(IDEMPOTENT_REPLAYED, Boolean.toString(result.replayed()))
                .body(RideResponse.from(result.ride()));
    }

    @GetMapping("/{rideId}")
    public RideResponse get(@PathVariable UUID rideId) {
        return RideResponse.from(rides.getForRead(rideId));
    }

    @PostMapping("/{rideId}/cancel")
    public RideResponse cancel(@PathVariable UUID rideId) {
        return RideResponse.from(commands.cancel(rideId));
    }

    @PostMapping("/{rideId}/complete")
    public RideResponse complete(@PathVariable UUID rideId) {
        return RideResponse.from(commands.complete(rideId));
    }

    @GetMapping("/{rideId}/history")
    public List<RideHistoryEntry> history(@PathVariable UUID rideId) {
        return rides.history(rideId);
    }
}
