package com.ridehailing.driver.frameworks;

import com.ridehailing.common.events.MessagingConfig;
import com.ridehailing.contracts.Topics;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;
import org.springframework.kafka.core.KafkaAdmin;

@Configuration(proxyBeanMethods = false)
public class DriverMessagingConfig {

    /** Driver owns and alone publishes to driver-events. */
    @Bean
    KafkaAdmin.NewTopics driverEventsTopics() {
        return MessagingConfig.ownedTopic(Topics.DRIVER_EVENTS);
    }
}
