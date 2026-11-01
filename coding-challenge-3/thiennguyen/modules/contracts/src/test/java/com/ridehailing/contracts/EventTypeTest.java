package com.ridehailing.contracts;

import static org.assertj.core.api.Assertions.assertThat;

import java.util.Arrays;
import org.junit.jupiter.api.Test;

class EventTypeTest {

    @Test
    void wireNamesAreUniqueAndResolveBackToTheirType() {
        assertThat(Arrays.stream(EventType.values()).map(EventType::wireName).distinct()).hasSize(EventType.values().length);
        for (EventType type : EventType.values()) {
            assertThat(EventType.fromWireName(type.wireName())).contains(type);
        }
    }

    @Test
    void everyEventBelongsToExactlyOneOwnedTopic() {
        assertThat(EventType.values()).allSatisfy(type -> assertThat(type.topic())
                .isIn(Topics.RIDE_EVENTS, Topics.DISPATCH_EVENTS, Topics.DRIVER_EVENTS));
        assertThat(EventType.RIDE_REQUESTED.topic()).isEqualTo(Topics.RIDE_EVENTS);
        assertThat(EventType.RIDE_OFFERED.topic()).isEqualTo(Topics.DISPATCH_EVENTS);
        assertThat(EventType.DRIVER_AVAILABILITY_CHANGED.topic()).isEqualTo(Topics.DRIVER_EVENTS);
    }

    @Test
    void payloadTypesAreVersionedRecords() {
        assertThat(EventType.values()).allSatisfy(type -> {
            assertThat(type.payloadType().isRecord()).isTrue();
            assertThat(type.payloadType().getSimpleName()).endsWith("V1");
        });
    }

    @Test
    void unknownWireNamesAreNotResolved() {
        assertThat(EventType.fromWireName("SomethingElse")).isEmpty();
    }
}
