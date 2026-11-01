package com.ridehailing.common.logging;

import static org.assertj.core.api.Assertions.assertThat;

import org.junit.jupiter.params.ParameterizedTest;
import org.junit.jupiter.params.provider.CsvSource;
import org.springframework.boot.ansi.AnsiColor;

class LayerColorConverterTest {

    @ParameterizedTest(name = "{0} → {1}")
    @CsvSource({
            "com.ridehailing.ride.adapters.web.RideController, BLUE",
            "com.ridehailing.ride.adapters.web.RideResponse, BLUE",
            "com.ridehailing.dispatch.adapters.messaging.DispatchEventListener, CYAN",
            "com.ridehailing.dispatch.application.DispatchService, BRIGHT_MAGENTA",
            "com.ridehailing.ride.application.port.RideRepository, BRIGHT_MAGENTA",
            "com.ridehailing.dispatch.domain.matching.NearestDriverMatcher, BRIGHT_MAGENTA",
            "com.ridehailing.ride.adapters.persistence.JdbcRideRepository, MAGENTA",
            "com.ridehailing.driver.adapters.cache.RedisDriverReadCache, MAGENTA",
            "com.ridehailing.driver.frameworks.DriverMessagingConfig, BRIGHT_BLUE",
            "com.ridehailing.common.events.OutboxWriter, BRIGHT_BLUE",
            "com.ridehailing.RideHailingApplication, DEFAULT",
            "org.springframework.boot.SpringApplication, BRIGHT_BLACK",
            "com.zaxxer.hikari.HikariDataSource, BRIGHT_BLACK"
    })
    void colorsEachClassByItsLayer(String loggerName, AnsiColor expected) {
        assertThat(LayerColorConverter.colorOf(loggerName)).isEqualTo(expected);
    }
}
