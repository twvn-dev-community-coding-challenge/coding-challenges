package com.ridehailing.common.events;

import com.ridehailing.contracts.Topics;
import io.micrometer.core.instrument.MeterRegistry;
import org.apache.kafka.clients.admin.NewTopic;
import org.apache.kafka.common.TopicPartition;
import org.apache.kafka.common.config.TopicConfig;
import org.springframework.boot.context.properties.EnableConfigurationProperties;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;
import org.springframework.kafka.config.TopicBuilder;
import org.springframework.kafka.core.KafkaAdmin;
import org.springframework.kafka.core.KafkaTemplate;
import org.springframework.kafka.listener.ConsumerAwareRecordRecoverer;
import org.springframework.kafka.listener.DeadLetterPublishingRecoverer;
import org.springframework.kafka.listener.DefaultErrorHandler;
import org.springframework.kafka.support.ExponentialBackOffWithMaxRetries;

@Configuration(proxyBeanMethods = false)
@EnableConfigurationProperties({OutboxProperties.class, EventRetentionProperties.class})
public class MessagingConfig {

    /**
     * A module's own topic plus its dead-letter topic, both kept for {@link Topics#RETENTION_DAYS}; each owning module
     * declares one of these as a bean.
     */
    public static KafkaAdmin.NewTopics ownedTopic(String topic) {
        return new KafkaAdmin.NewTopics(newTopic(topic), newTopic(Topics.deadLetter(topic)));
    }

    /**
     * Retries, then dead-letters; {@code ridehailing.events.dead.letters{topic}} counts each record dead-lettered.
     */
    @Bean
    DefaultErrorHandler kafkaErrorHandler(KafkaTemplate<?, ?> kafkaTemplate, MeterRegistry meters) {
        DeadLetterPublishingRecoverer deadLetters = new DeadLetterPublishingRecoverer(kafkaTemplate,
                (failedRecord, exception) -> new TopicPartition(
                        Topics.deadLetter(failedRecord.topic()), failedRecord.partition()));
        ConsumerAwareRecordRecoverer recoverer = (failedRecord, consumer, exception) -> {
            deadLetters.accept(failedRecord, consumer, exception);
            meters.counter("ridehailing.events.dead.letters", "topic", failedRecord.topic()).increment();
        };
        ExponentialBackOffWithMaxRetries backOff = new ExponentialBackOffWithMaxRetries(Topics.MAX_RETRIES);
        backOff.setInitialInterval(500);
        backOff.setMultiplier(2.0);
        DefaultErrorHandler handler = new DefaultErrorHandler(recoverer, backOff);
        handler.addNotRetryableExceptions(MalformedEventException.class);
        return handler;
    }

    private static NewTopic newTopic(String name) {
        return TopicBuilder.name(name)
                .partitions(Topics.PARTITIONS)
                .replicas(1)
                .config(TopicConfig.RETENTION_MS_CONFIG, Long.toString(Topics.RETENTION_DAYS.toMillis()))
                .build();
    }
}
