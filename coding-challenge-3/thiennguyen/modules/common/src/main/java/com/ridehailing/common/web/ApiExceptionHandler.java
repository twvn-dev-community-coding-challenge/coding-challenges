package com.ridehailing.common.web;

import com.ridehailing.common.error.ApplicationException;
import com.ridehailing.common.error.ConflictException;
import com.ridehailing.common.error.NotFoundException;
import com.ridehailing.common.error.UnprocessableException;
import com.ridehailing.common.error.ValidationException;
import com.ridehailing.common.kernel.InvalidValueException;
import java.util.List;
import java.util.Map;
import java.util.Objects;
import lombok.extern.slf4j.Slf4j;
import org.springframework.dao.DataIntegrityViolationException;
import org.springframework.http.HttpHeaders;
import org.springframework.http.HttpStatus;
import org.springframework.http.HttpStatusCode;
import org.springframework.http.ProblemDetail;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.MethodArgumentNotValidException;
import org.springframework.web.bind.annotation.ExceptionHandler;
import org.springframework.web.bind.annotation.RestControllerAdvice;
import org.springframework.web.context.request.WebRequest;
import org.springframework.web.servlet.mvc.method.annotation.ResponseEntityExceptionHandler;

/**
 * Maps every error the API can return to an RFC 9457 problem body carrying a stable {@code code}. The HTTP status of an
 * {@link ApplicationException} is decided here, in the web adapter, not by the code that throws it.
 */
@RestControllerAdvice
@Slf4j
public class ApiExceptionHandler extends ResponseEntityExceptionHandler {

    /** The sealed hierarchy makes this switch exhaustive: a new kind of exception does not compile without a status. */
    static HttpStatus statusOf(ApplicationException exception) {
        return switch (exception) {
            case NotFoundException _ -> HttpStatus.NOT_FOUND;
            case ConflictException _ -> HttpStatus.CONFLICT;
            case ValidationException _ -> HttpStatus.BAD_REQUEST;
            case UnprocessableException _ -> HttpStatus.UNPROCESSABLE_CONTENT;
        };
    }

    @ExceptionHandler(ApplicationException.class)
    ResponseEntity<ProblemDetail> handleApplicationException(ApplicationException exception) {
        HttpStatus status = statusOf(exception);
        log.info("Request rejected with {}: {}", exception.code(), exception.getMessage());
        ProblemDetail problem = ProblemDetail.forStatusAndDetail(status, exception.getMessage());
        problem.setTitle(status.getReasonPhrase());
        problem.setProperty("code", exception.code());
        exception.details().forEach(problem::setProperty);
        return ResponseEntity.status(status).body(problem);
    }

    @ExceptionHandler(InvalidValueException.class)
    ResponseEntity<ProblemDetail> handleInvalidValue(InvalidValueException exception) {
        return handleApplicationException(new ValidationException(exception.code(), exception.getMessage()));
    }

    /** A constraint backstop fired: another request changed the same rows first. */
    @ExceptionHandler(DataIntegrityViolationException.class)
    ResponseEntity<ProblemDetail> handleConstraintViolation(DataIntegrityViolationException exception) {
        log.warn("Concurrent update rejected by a database constraint", exception);
        ProblemDetail problem = ProblemDetail.forStatusAndDetail(
                HttpStatus.CONFLICT, "The resource was changed by a concurrent request; retry with current state");
        problem.setTitle(HttpStatus.CONFLICT.getReasonPhrase());
        problem.setProperty("code", "CONCURRENT_UPDATE");
        return ResponseEntity.status(HttpStatus.CONFLICT).body(problem);
    }

    @Override
    protected ResponseEntity<Object> handleMethodArgumentNotValid(
            MethodArgumentNotValidException exception, HttpHeaders headers, HttpStatusCode status, WebRequest request) {
        ProblemDetail problem = exception.getBody();
        problem.setProperty("code", "VALIDATION_FAILED");
        List<Map<String, String>> errors = exception.getBindingResult().getFieldErrors().stream()
                .map(error -> Map.of(
                        "field", error.getField(),
                        "message", Objects.requireNonNullElse(error.getDefaultMessage(), "is invalid")))
                .toList();
        problem.setProperty("errors", errors);
        return handleExceptionInternal(exception, problem, headers, status, request);
    }
}
