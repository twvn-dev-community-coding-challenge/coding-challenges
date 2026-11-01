package com.ridehailing.ride.adapters.web;

import static org.hamcrest.Matchers.nullValue;
import static org.mockito.ArgumentMatchers.any;
import static org.mockito.ArgumentMatchers.eq;
import static org.mockito.BDDMockito.given;
import static org.mockito.Mockito.verifyNoInteractions;
import static org.springframework.test.web.servlet.request.MockMvcRequestBuilders.get;
import static org.springframework.test.web.servlet.request.MockMvcRequestBuilders.post;
import static org.springframework.test.web.servlet.result.MockMvcResultMatchers.header;
import static org.springframework.test.web.servlet.result.MockMvcResultMatchers.jsonPath;
import static org.springframework.test.web.servlet.result.MockMvcResultMatchers.status;

import com.ridehailing.common.kernel.geo.Location;
import com.ridehailing.ride.application.CreateRideCommand;
import com.ridehailing.ride.application.RideCommandService;
import com.ridehailing.ride.application.RideCreationService;
import com.ridehailing.ride.application.RideLifecycle;
import com.ridehailing.ride.domain.Ride;
import java.time.Instant;
import java.util.UUID;
import org.junit.jupiter.api.Test;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.boot.webmvc.test.autoconfigure.WebMvcTest;
import org.springframework.http.MediaType;
import org.springframework.test.context.bean.override.mockito.MockitoBean;
import org.springframework.test.web.servlet.MockMvc;

@WebMvcTest(RideController.class)
class RideControllerTest {

    private static final String VALID_BODY = """
            {"passengerId": "P001",
             "pickup": {"latitude": 10.7769, "longitude": 106.7009},
             "destination": {"latitude": 10.8231, "longitude": 106.6297}}
            """;

    @Autowired
    private MockMvc mvc;

    @MockitoBean
    private RideCreationService creation;

    @MockitoBean
    private RideCommandService commands;

    @MockitoBean
    private RideLifecycle rides;

    @Test
    void createsARideAndReturns201WithLocation() throws Exception {
        Ride ride = ride();
        given(creation.create(eq("key-1"), any(CreateRideCommand.class)))
                .willReturn(new RideCreationService.CreationResult(ride, false));

        mvc.perform(post("/rides").header("Idempotency-Key", "key-1")
                        .contentType(MediaType.APPLICATION_JSON).content(VALID_BODY))
                .andExpect(status().isCreated())
                .andExpect(header().string("Location", "/rides/" + ride.id()))
                .andExpect(header().string("Idempotent-Replayed", "false"))
                .andExpect(jsonPath("$.id").value(ride.id().toString()))
                .andExpect(jsonPath("$.status").value("REQUESTED"))
                .andExpect(jsonPath("$.assignedDriverId").value(nullValue()));
    }

    @Test
    void marksAReplayedResponse() throws Exception {
        given(creation.create(eq("key-1"), any(CreateRideCommand.class)))
                .willReturn(new RideCreationService.CreationResult(ride(), true));

        mvc.perform(post("/rides").header("Idempotency-Key", "key-1")
                        .contentType(MediaType.APPLICATION_JSON).content(VALID_BODY))
                .andExpect(status().isCreated())
                .andExpect(header().string("Idempotent-Replayed", "true"));
    }

    @Test
    void requiresAnIdempotencyKey() throws Exception {
        mvc.perform(post("/rides").contentType(MediaType.APPLICATION_JSON).content(VALID_BODY))
                .andExpect(status().isBadRequest())
                .andExpect(jsonPath("$.code").value("IDEMPOTENCY_KEY_REQUIRED"));
        verifyNoInteractions(creation);
    }

    @Test
    void rejectsAnOverlongIdempotencyKey() throws Exception {
        mvc.perform(post("/rides").header("Idempotency-Key", "k".repeat(129))
                        .contentType(MediaType.APPLICATION_JSON).content(VALID_BODY))
                .andExpect(status().isBadRequest())
                .andExpect(jsonPath("$.code").value("IDEMPOTENCY_KEY_TOO_LONG"));
    }

    @Test
    void rejectsCoordinatesOutOfRange() throws Exception {
        String body = VALID_BODY.replace("10.7769", "91");
        mvc.perform(post("/rides").header("Idempotency-Key", "key-2")
                        .contentType(MediaType.APPLICATION_JSON).content(body))
                .andExpect(status().isBadRequest())
                .andExpect(jsonPath("$.code").value("VALIDATION_FAILED"))
                .andExpect(jsonPath("$.errors[0].field").value("pickup.latitude"));
        verifyNoInteractions(creation);
    }

    @Test
    void rejectsAMissingPassengerAndDestination() throws Exception {
        mvc.perform(post("/rides").header("Idempotency-Key", "key-3")
                        .contentType(MediaType.APPLICATION_JSON)
                        .content("{\"pickup\": {\"latitude\": 1, \"longitude\": 1}}"))
                .andExpect(status().isBadRequest())
                .andExpect(jsonPath("$.code").value("VALIDATION_FAILED"));
    }

    @Test
    void returns404ForAnUnknownRide() throws Exception {
        UUID rideId = UUID.randomUUID();
        given(rides.getForRead(rideId)).willThrow(RideLifecycle.notFound(rideId));

        mvc.perform(get("/rides/{id}", rideId))
                .andExpect(status().isNotFound())
                .andExpect(jsonPath("$.code").value("RIDE_NOT_FOUND"));
    }

    @Test
    void returns400ForAMalformedRideId() throws Exception {
        mvc.perform(get("/rides/not-a-uuid")).andExpect(status().isBadRequest());
    }

    private static Ride ride() {
        Instant now = Instant.parse("2026-10-26T10:00:00Z");
        return Ride.requested(UUID.randomUUID(), "P001", new Location(10.7769, 106.7009),
                new Location(10.8231, 106.6297), now);
    }
}
