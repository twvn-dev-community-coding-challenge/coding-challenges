package com.ridehailing.dispatch.application;

import com.ridehailing.common.logging.LogContext;
import com.ridehailing.dispatch.application.port.DispatchRequestRepository;
import java.time.Clock;
import java.util.UUID;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.scheduling.annotation.Scheduled;
import org.springframework.stereotype.Component;

/**
 * Self-healing for rides still waiting for a driver: retries matching for requests nobody has tried recently, which
 * covers dead-lettered or lost messages and drivers who came online after a ride found no candidate. Reads only
 * dispatch's own table.
 */
@Component
@RequiredArgsConstructor
@Slf4j
public class MatchingSweeper {

    private final DispatchRequestRepository requests;
    private final DispatchService dispatch;
    private final DispatchProperties properties;
    private final Clock clock;

    @Scheduled(fixedDelayString = "${dispatch.sweeper-interval-ms:5000}")
    public void sweep() {
        for (UUID rideId : requests.findStaleWaiting(
                clock.instant().minus(properties.sweeperStaleAfter()), properties.sweeperBatchSize())) {
            try (LogContext ignored = LogContext.ride(rideId)) {
                log.info("Sweeper retrying dispatch for a ride still waiting for a driver");
                dispatch.dispatch(rideId);
            } catch (RuntimeException exception) {
                log.warn("Sweeper dispatch failed: {}", exception.toString());
            }
        }
    }
}
