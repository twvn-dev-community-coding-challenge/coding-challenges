package com.ridehailing.integration;

import static org.assertj.core.api.Assertions.assertThat;

import com.ridehailing.driver.domain.DriverStatus;
import com.ridehailing.support.IntegrationTestSupport;
import com.ridehailing.support.TestCity;
import java.util.List;
import java.util.UUID;
import org.junit.jupiter.api.Tag;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.springframework.boot.test.system.CapturedOutput;
import org.springframework.boot.test.system.OutputCaptureExtension;

/** "What happened to ride R001?" must be answerable from the logs alone: every line carries its ride, driver and event. */
@Tag("observability")
@ExtendWith(OutputCaptureExtension.class)
class ObservabilityIT extends IntegrationTestSupport {

    private final TestCity city = TestCity.next();

    @Test
    void everyStepOfARideIsLoggedWithItsIdentifiers(CapturedOutput output) {
        String driver = fixtures.driver(DriverStatus.AVAILABLE, city.kmNorth(0.5));
        UUID rideId = createRide(fixtures.passenger(), city.centre()).id();
        awaitOfferedTo(rideId, driver);
        acceptAndAwait(rideId, driver);

        eventually().untilAsserted(() -> {
            List<String> rideLines = output.getOut().lines()
                    .filter(line -> line.contains("rideId=" + rideId))
                    .toList();
            assertThat(rideLines).anyMatch(line -> line.contains("Ride created in REQUESTED"));
            assertThat(rideLines).anyMatch(line -> line.contains("dispatch consuming RideRequested")
                    && hasEventId(line));
            assertThat(rideLines).anyMatch(line -> line.contains("ride consuming RideOffered")
                    && hasEventId(line));
            assertThat(rideLines).anyMatch(line -> line.contains("Ride offered to driver")
                    && line.contains("driverId=" + driver));
            assertThat(rideLines).anyMatch(line -> line.contains("DRIVER_ACCEPTED (driver accepted)"));
        });
    }

    private static boolean hasEventId(String line) {
        return line.contains("eventId=") && !line.contains("eventId=-");
    }
}
