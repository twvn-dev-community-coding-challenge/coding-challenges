package com.ridehailing.common.kernel.geo;

import static org.assertj.core.api.Assertions.assertThatCode;
import static org.assertj.core.api.Assertions.assertThatThrownBy;

import com.ridehailing.common.kernel.InvalidValueException;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.params.ParameterizedTest;
import org.junit.jupiter.params.provider.CsvSource;

class LocationTest {

    @ParameterizedTest
    @CsvSource({"-90, -180", "90, 180", "0, 0", "10.7769, 106.7009"})
    void acceptsCoordinatesWithinRangeIncludingBounds(double latitude, double longitude) {
        assertThatCode(() -> new Location(latitude, longitude)).doesNotThrowAnyException();
    }

    @ParameterizedTest
    @CsvSource({"90.0001, 0", "-90.0001, 0", "0, 180.0001", "0, -180.0001", "91, 106.7", "10.7, -181"})
    void rejectsCoordinatesOutOfRange(double latitude, double longitude) {
        assertThatThrownBy(() -> new Location(latitude, longitude)).isInstanceOf(InvalidValueException.class);
    }

    @Test
    void rejectsNonFiniteCoordinates() {
        assertThatThrownBy(() -> new Location(Double.NaN, 0)).isInstanceOf(InvalidValueException.class);
        assertThatThrownBy(() -> new Location(0, Double.POSITIVE_INFINITY)).isInstanceOf(InvalidValueException.class);
    }
}
