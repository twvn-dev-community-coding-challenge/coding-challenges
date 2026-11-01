package com.ridehailing.dispatch;

import static org.assertj.core.api.Assertions.assertThat;

import com.ridehailing.TestcontainersConfiguration;
import com.ridehailing.common.events.EventCodec;
import com.ridehailing.contracts.EventEnvelope;
import com.ridehailing.contracts.EventType;
import com.ridehailing.contracts.GeoPoint;
import com.ridehailing.contracts.RideRequestedV1;
import com.ridehailing.contracts.Topics;
import com.ridehailing.dispatch.application.OfferService;
import java.time.Duration;
import java.time.Instant;
import java.time.OffsetDateTime;
import java.time.ZoneOffset;
import java.util.Arrays;
import java.util.List;
import java.util.Objects;
import java.util.UUID;
import java.util.concurrent.TimeUnit;
import org.awaitility.Awaitility;
import org.junit.jupiter.api.Test;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.context.ApplicationContext;
import org.springframework.context.annotation.Import;
import org.springframework.jdbc.core.simple.JdbcClient;
import org.springframework.kafka.core.KafkaTemplate;
import org.springframework.modulith.test.ApplicationModuleTest.BootstrapMode;
import org.springframework.modulith.test.ApplicationModuleTest;

/**
 * Proves dispatch can be extracted: Spring Modulith boots the dispatch module on its own, with only the shared
 * {@code common} and {@code contracts} modules, and dispatch serves a ride from nothing but a RideRequested event and
 * its own supply table. The database still holds every module's tables because Flyway migrates them all; a split-out
 * service would migrate only {@code db/migration/common} and {@code db/migration/dispatch}.
 *
 * <p>{@code verifyAutomatically} is off: on the integration-test classpath Modulith would also treat the test support
 * packages as modules. Boundaries are verified on production classes by {@code ModularityTest}.
 */
@ApplicationModuleTest(mode = BootstrapMode.STANDALONE, extraIncludes = {"common", "contracts"},
        verifyAutomatically = false)
@Import(TestcontainersConfiguration.class)
class DispatchModuleIT {

    private static final List<String> OTHER_MODULES =
            List.of("com.ridehailing.ride.", "com.ridehailing.driver.");

    private final ApplicationContext context;
    private final KafkaTemplate<String, String> kafka;
    private final EventCodec codec;
    private final JdbcClient jdbc;

    @Autowired
    DispatchModuleIT(ApplicationContext context, KafkaTemplate<String, String> kafka, EventCodec codec, JdbcClient jdbc) {
        this.context = context;
        this.kafka = kafka;
        this.codec = codec;
        this.jdbc = jdbc;
    }

    @Test
    void bootsWithoutAnyBeanFromRideOrDriver() {
        assertThat(context.getBeanNamesForType(OfferService.class)).hasSize(1);
        List<String> foreign = Arrays.stream(context.getBeanDefinitionNames())
                .map(name -> context.getType(name, false))
                .filter(Objects::nonNull)
                .map(Class::getName)
                .filter(type -> OTHER_MODULES.stream().anyMatch(type::startsWith))
                .toList();
        assertThat(foreign).as("beans from other business modules").isEmpty();
    }

    @Test
    void offersARideFromARideRequestedEventAlone() throws Exception {
        String driverId = "D-STANDALONE-" + UUID.randomUUID().toString().substring(0, 8);
        GeoPoint pickup = new GeoPoint(-45.0, 170.0);
        jdbc.sql("""
                        INSERT INTO driver_supply (driver_id, online, latitude, longitude, updated_at)
                        VALUES (:id, TRUE, :latitude, :longitude, :now)
                        """)
                .param("id", driverId)
                .param("latitude", pickup.latitude() + 0.5 / 111.195)
                .param("longitude", pickup.longitude())
                .param("now", OffsetDateTime.now(ZoneOffset.UTC))
                .update();
        UUID rideId = UUID.randomUUID();

        EventEnvelope requested = EventEnvelope.create(EventType.RIDE_REQUESTED, rideId.toString(), null,
                Instant.now(), codec.toPayload(new RideRequestedV1(rideId, "P-STANDALONE", pickup)));
        kafka.send(Topics.RIDE_EVENTS, rideId.toString(), codec.encode(requested)).get(10, TimeUnit.SECONDS);

        Awaitility.await().atMost(Duration.ofSeconds(60)).pollInterval(Duration.ofMillis(200)).untilAsserted(() ->
                assertThat(jdbc.sql("SELECT status || ':' || COALESCE(driver_id, '') FROM dispatch_requests WHERE ride_id = :id")
                        .param("id", rideId)
                        .query(String.class)
                        .optional())
                        .contains("OFFERED:" + driverId));
        assertThat(jdbc.sql("SELECT count(*) FROM outbox WHERE aggregate_id = :id AND event_type = 'RideOffered'")
                .param("id", rideId.toString())
                .query(Long.class)
                .single()).isEqualTo(1);
    }
}
