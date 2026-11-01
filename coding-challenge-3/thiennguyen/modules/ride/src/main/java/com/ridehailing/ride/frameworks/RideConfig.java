package com.ridehailing.ride.frameworks;

import com.ridehailing.ride.application.RideProperties;
import org.springframework.boot.context.properties.EnableConfigurationProperties;
import org.springframework.context.annotation.Configuration;

@Configuration(proxyBeanMethods = false)
@EnableConfigurationProperties(RideProperties.class)
public class RideConfig {
}
