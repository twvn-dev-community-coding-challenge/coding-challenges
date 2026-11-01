package com.ridehailing.integration;

import static org.assertj.core.api.Assertions.assertThat;

import com.ridehailing.contracts.Topics;
import com.ridehailing.support.IntegrationTestSupport;
import java.util.List;
import java.util.Map;
import java.util.concurrent.TimeUnit;
import org.apache.kafka.clients.admin.AdminClient;
import org.apache.kafka.clients.admin.Config;
import org.apache.kafka.clients.admin.ConfigEntry;
import org.apache.kafka.common.config.ConfigResource;
import org.apache.kafka.common.config.TopicConfig;
import org.junit.jupiter.api.Test;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.kafka.core.KafkaAdmin;

/** Every topic and dead-letter topic is created with the retention that event de-duplication relies on. */
class KafkaTopicsIT extends IntegrationTestSupport {

    @Autowired
    private KafkaAdmin kafkaAdmin;

    @Test
    void everyTopicAndDeadLetterTopicKeepsRecordsForTheTopicRetention() throws Exception {
        List<ConfigResource> topics = List.of(Topics.RIDE_EVENTS, Topics.DISPATCH_EVENTS, Topics.DRIVER_EVENTS).stream()
                .flatMap(topic -> List.of(topic, Topics.deadLetter(topic)).stream())
                .map(topic -> new ConfigResource(ConfigResource.Type.TOPIC, topic))
                .toList();

        try (AdminClient admin = AdminClient.create(kafkaAdmin.getConfigurationProperties())) {
            Map<ConfigResource, Config> configs = admin.describeConfigs(topics).all().get(10, TimeUnit.SECONDS);

            assertThat(configs).hasSize(6);
            configs.forEach((topic, config) -> {
                ConfigEntry retention = config.get(TopicConfig.RETENTION_MS_CONFIG);
                // Set on the topic itself, not inherited from the broker default (which is also 7 days).
                assertThat(retention.source()).as(topic.name()).isEqualTo(ConfigEntry.ConfigSource.DYNAMIC_TOPIC_CONFIG);
                assertThat(retention.value()).as(topic.name())
                        .isEqualTo(Long.toString(Topics.RETENTION_DAYS.toMillis()));
            });
        }
    }
}
