package com.ridehailing.common.events;

import static org.assertj.core.api.Assertions.assertThatCode;
import static org.assertj.core.api.Assertions.assertThatThrownBy;

import com.ridehailing.contracts.Topics;
import java.time.Duration;
import org.junit.jupiter.api.Test;

class EventRetentionPropertiesTest {

    @Test
    void acceptsMarkersKeptLongerThanTheTopics() {
        assertThatCode(() -> new EventRetentionProperties(Duration.ofDays(1), Topics.RETENTION_DAYS.plusDays(1), 1000))
                .doesNotThrowAnyException();
    }

    @Test
    void rejectsMarkersKeptNoLongerThanTheTopics() {
        assertThatThrownBy(() -> new EventRetentionProperties(Duration.ofDays(1), Topics.RETENTION_DAYS, 1000))
                .isInstanceOf(IllegalArgumentException.class)
                .hasMessageContaining("processed-events");
    }

    @Test
    void rejectsANonPositiveOutboxRetentionOrBatchSize() {
        assertThatThrownBy(() -> new EventRetentionProperties(Duration.ZERO, Duration.ofDays(14), 1000))
                .isInstanceOf(IllegalArgumentException.class);
        assertThatThrownBy(() -> new EventRetentionProperties(Duration.ofDays(1), Duration.ofDays(14), 0))
                .isInstanceOf(IllegalArgumentException.class);
    }
}
