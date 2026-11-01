package com.ridehailing.ride;

import org.springframework.boot.autoconfigure.SpringBootApplication;

/**
 * The ride module has no application class of its own (that lives in :app); slice tests such as
 * {@code @WebMvcTest} find this one instead. It also scans {@code common} for the error handler and logging filter.
 */
@SpringBootApplication(scanBasePackages = {"com.ridehailing.ride", "com.ridehailing.common"})
class RideTestApplication {
}
