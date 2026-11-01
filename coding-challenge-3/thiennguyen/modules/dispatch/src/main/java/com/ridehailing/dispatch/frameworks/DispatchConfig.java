package com.ridehailing.dispatch.frameworks;

import com.ridehailing.dispatch.application.DispatchProperties;
import com.ridehailing.dispatch.domain.matching.NearestDriverMatcher;
import org.springframework.boot.context.properties.EnableConfigurationProperties;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;

@Configuration(proxyBeanMethods = false)
@EnableConfigurationProperties(DispatchProperties.class)
public class DispatchConfig {

    @Bean
    NearestDriverMatcher nearestDriverMatcher(DispatchProperties properties) {
        return new NearestDriverMatcher(properties.maxRadiusKm());
    }
}
