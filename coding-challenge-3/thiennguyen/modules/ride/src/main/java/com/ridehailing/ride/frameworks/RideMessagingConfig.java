package com.ridehailing.ride.frameworks;

import com.ridehailing.common.events.MessagingConfig;
import com.ridehailing.contracts.Topics;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;
import org.springframework.kafka.core.KafkaAdmin;

@Configuration(proxyBeanMethods = false)
public class RideMessagingConfig {
    @Bean
    KafkaAdmin.NewTopics rideEventsTopics() {
        return MessagingConfig.ownedTopic(Topics.RIDE_EVENTS);
    }
}
