package com.ridehailing.common.web;

import static org.assertj.core.api.Assertions.assertThat;

import com.ridehailing.common.error.ApplicationException;
import com.ridehailing.common.error.ConflictException;
import com.ridehailing.common.error.NotFoundException;
import com.ridehailing.common.error.UnprocessableException;
import com.ridehailing.common.error.ValidationException;
import com.ridehailing.common.kernel.InvalidValueException;
import java.util.Map;
import java.util.stream.Stream;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.params.ParameterizedTest;
import org.junit.jupiter.params.provider.Arguments;
import org.junit.jupiter.params.provider.MethodSource;
import org.springframework.http.HttpStatus;
import org.springframework.http.ProblemDetail;
import org.springframework.http.ResponseEntity;

class ApiExceptionHandlerTest {

    private final ApiExceptionHandler handler = new ApiExceptionHandler();

    static Stream<Arguments> exceptions() {
        return Stream.of(
                Arguments.of(new NotFoundException("RIDE_NOT_FOUND", "no such ride"), HttpStatus.NOT_FOUND),
                Arguments.of(new ConflictException("INVALID_TRANSITION", "not now"), HttpStatus.CONFLICT),
                Arguments.of(new ValidationException("UNKNOWN_PASSENGER", "who?"), HttpStatus.BAD_REQUEST),
                Arguments.of(new UnprocessableException("IDEMPOTENCY_KEY_REUSED", "other body"),
                        HttpStatus.UNPROCESSABLE_CONTENT));
    }

    @ParameterizedTest(name = "{0} → {1}")
    @MethodSource("exceptions")
    void theWebAdapterDecidesEachStatus(ApplicationException exception, HttpStatus expected) {
        ResponseEntity<ProblemDetail> response = handler.handleApplicationException(exception);

        assertThat(response.getStatusCode()).isEqualTo(expected);
        assertThat(response.getBody().getStatus()).isEqualTo(expected.value());
        assertThat(response.getBody().getTitle()).isEqualTo(expected.getReasonPhrase());
        assertThat(response.getBody().getDetail()).isEqualTo(exception.getMessage());
        assertThat(response.getBody().getProperties()).containsEntry("code", exception.code());
    }

    @Test
    void detailsBecomeProblemProperties() {
        ConflictException exception =
                new ConflictException("INVALID_TRANSITION", "not now", Map.of("currentStatus", "COMPLETED"));

        ProblemDetail problem = handler.handleApplicationException(exception).getBody();

        assertThat(problem.getProperties()).containsEntry("currentStatus", "COMPLETED");
    }

    @Test
    void anInvalidKernelValueIsABadRequest() {
        ResponseEntity<ProblemDetail> response =
                handler.handleInvalidValue(new InvalidValueException("INVALID_LOCATION", "latitude out of range"));

        assertThat(response.getStatusCode()).isEqualTo(HttpStatus.BAD_REQUEST);
        assertThat(response.getBody().getProperties()).containsEntry("code", "INVALID_LOCATION");
    }
}
