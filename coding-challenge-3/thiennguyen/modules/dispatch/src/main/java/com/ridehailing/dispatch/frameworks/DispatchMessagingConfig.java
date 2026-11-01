package com.ridehailing.dispatch.frameworks;

import com.ridehailing.common.events.MessagingConfig;
import com.ridehailing.contracts.Topics;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;
import org.springframework.kafka.core.KafkaAdmin;

@Configuration(proxyBeanMethods = false)
public class DispatchMessagingConfig {

    /**
     * Dispatch owns and alone publishes to dispatch-events.
     */
    @Bean
    KafkaAdmin.NewTopics dispatchEventsTopics() {
        return MessagingConfig.ownedTopic(Topics.DISPATCH_EVENTS);
    }
}
