package com.ridehailing.common.logging;

import jakarta.servlet.FilterChain;
import jakarta.servlet.ServletException;
import jakarta.servlet.http.HttpServletRequest;
import jakarta.servlet.http.HttpServletResponse;
import java.io.IOException;
import java.util.regex.Matcher;
import java.util.regex.Pattern;
import lombok.extern.slf4j.Slf4j;
import org.springframework.core.Ordered;
import org.springframework.core.annotation.Order;
import org.springframework.stereotype.Component;
import org.springframework.web.filter.OncePerRequestFilter;

/** Tags every log line of a request with the ride and driver identifiers found in its path. */
@Component
@Order(Ordered.HIGHEST_PRECEDENCE)
@Slf4j
public class CorrelationFilter extends OncePerRequestFilter {

    private static final Pattern RIDE_PATH = Pattern.compile("^/rides/([^/]+)");
    private static final Pattern DRIVER_PATH = Pattern.compile("^/drivers/([^/]+)|/offers/([^/]+)");

    @Override
    protected void doFilterInternal(HttpServletRequest request, HttpServletResponse response, FilterChain chain)
            throws ServletException, IOException {
        String path = request.getRequestURI();
        try (LogContext ignored = LogContext.with(LogContext.RIDE_ID, rideId(path))
                .and(LogContext.DRIVER_ID, driverId(path))) {
            chain.doFilter(request, response);
            log.info("{} {} -> {}", request.getMethod(), path, response.getStatus());
        }
    }

    private static String rideId(String path) {
        Matcher matcher = RIDE_PATH.matcher(path);
        return matcher.find() ? matcher.group(1) : null;
    }

    private static String driverId(String path) {
        Matcher matcher = DRIVER_PATH.matcher(path);
        if (!matcher.find()) {
            return null;
        }
        return matcher.group(1) != null ? matcher.group(1) : matcher.group(2);
    }
}
