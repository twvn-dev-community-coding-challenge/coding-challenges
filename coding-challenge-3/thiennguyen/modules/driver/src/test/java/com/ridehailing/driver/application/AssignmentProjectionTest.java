package com.ridehailing.driver.application;

import static org.assertj.core.api.Assertions.assertThat;

import com.ridehailing.common.events.EventCodec;
import com.ridehailing.common.events.ProcessedEvents;
import com.ridehailing.common.kernel.geo.Location;
import com.ridehailing.contracts.AssignmentState;
import com.ridehailing.contracts.DriverAssignmentChangedV1;
import com.ridehailing.contracts.EventEnvelope;
import com.ridehailing.contracts.EventType;
import com.ridehailing.driver.application.port.DriverReadCache;
import com.ridehailing.driver.application.port.DriverRepository;
import com.ridehailing.driver.domain.Assignment;
import com.ridehailing.driver.domain.Driver;
import com.ridehailing.driver.domain.DriverAssignment;
import com.ridehailing.driver.domain.DriverStatus;
import com.ridehailing.driver.domain.DriverTransition;
import java.time.Instant;
import java.util.ArrayList;
import java.util.HashSet;
import java.util.List;
import java.util.Map;
import java.util.Optional;
import java.util.Set;
import java.util.UUID;
import org.junit.jupiter.api.Test;

/**
 * The use case runs on plain fakes of its ports and of {@code common}'s interfaces: no database, Kafka or Spring.
 */
class AssignmentProjectionTest {

    private static final Instant NOW = Instant.parse("2026-10-26T10:00:00Z");

    private final RecordingDrivers drivers = new RecordingDrivers();
    private final RecordingCache cache = new RecordingCache();
    private final Set<String> processed = new HashSet<>();
    private final ProcessedEvents processedEvents = (consumer, eventId) -> processed.add(consumer + ":" + eventId);
    private final FixedPayloadCodec codec = new FixedPayloadCodec();
    private final AssignmentProjection projection = new AssignmentProjection(drivers, cache, processedEvents, codec);

    @Test
    void anAssignmentIsAppliedAndTheDriverCachedAfterCommit() {
        UUID rideId = UUID.randomUUID();

        projection.apply(event(new DriverAssignmentChangedV1("D001", rideId, AssignmentState.OFFERED, 3)));

        assertThat(drivers.applied)
                .containsExactly(new Applied("D001", DriverAssignment.of(Assignment.OFFERED, rideId), 3));
        assertThat(cache.afterCommit).extracting(Driver::id).containsExactly("D001");
    }

    @Test
    void aDuplicateDeliveryChangesNothing() {
        EventEnvelope envelope =
                event(new DriverAssignmentChangedV1("D001", UUID.randomUUID(), AssignmentState.BUSY, 4));

        projection.apply(envelope);
        projection.apply(envelope);

        assertThat(drivers.applied).hasSize(1);
        assertThat(cache.afterCommit).hasSize(1);
    }

    @Test
    void anOlderAssignmentIsIgnoredAndNotCached() {
        drivers.newerAlreadyApplied = true;

        projection.apply(event(new DriverAssignmentChangedV1("D001", null, AssignmentState.NONE, 2)));

        assertThat(drivers.applied).hasSize(1);
        assertThat(cache.afterCommit).isEmpty();
    }

    private EventEnvelope event(DriverAssignmentChangedV1 change) {
        codec.payload = change;
        return EventEnvelope.create(EventType.DRIVER_ASSIGNMENT_CHANGED, null, change.driverId(), NOW, Map.of());
    }

    private record Applied(String driverId, DriverAssignment assignment, long version) {
    }

    private static final class RecordingDrivers implements DriverRepository {

        private final List<Applied> applied = new ArrayList<>();
        private boolean newerAlreadyApplied;

        @Override
        public Optional<Driver> applyAssignment(String driverId, DriverAssignment assignment, long assignmentVersion) {
            applied.add(new Applied(driverId, assignment, assignmentVersion));
            return newerAlreadyApplied ? Optional.empty()
                    : Optional.of(new Driver(driverId, DriverStatus.AVAILABLE, null, assignment, 1, NOW));
        }

        @Override
        public Optional<Driver> find(String driverId) {
            throw new UnsupportedOperationException();
        }

        @Override
        public Optional<Driver> findForUpdate(String driverId) {
            throw new UnsupportedOperationException();
        }

        @Override
        public Optional<Driver> applyAvailability(String driverId, DriverTransition transition) {
            throw new UnsupportedOperationException();
        }

        @Override
        public Optional<Driver> updateLocation(String driverId, Location location) {
            throw new UnsupportedOperationException();
        }
    }

    private static final class RecordingCache implements DriverReadCache {

        private final List<Driver> afterCommit = new ArrayList<>();

        @Override
        public void putAfterCommit(Driver driver) {
            afterCommit.add(driver);
        }

        @Override
        public Optional<Driver> get(String driverId) {
            throw new UnsupportedOperationException();
        }

        @Override
        public void put(Driver driver) {
            throw new UnsupportedOperationException();
        }
    }

    /** Hands the use case the payload the test set, instead of decoding JSON. */
    private static final class FixedPayloadCodec implements EventCodec {

        private Object payload;

        @Override
        public <T> T payload(EventEnvelope envelope, Class<T> payloadType) {
            return payloadType.cast(payload);
        }

        @Override
        public String encode(EventEnvelope envelope) {
            throw new UnsupportedOperationException();
        }

        @Override
        public Map<String, Object> toPayload(Object payloadRecord) {
            throw new UnsupportedOperationException();
        }

        @Override
        public EventEnvelope decode(String json) {
            throw new UnsupportedOperationException();
        }
    }
}
