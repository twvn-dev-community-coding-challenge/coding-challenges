package com.ridehailing.common.error;

import java.util.Map;

/**
 * A request the application refuses, with a stable {@code code} for clients. It says what went wrong, not how to
 * report it: the web adapter ({@code common.web.ApiExceptionHandler}) maps each kind to its HTTP status, so the
 * application stays free of HTTP. Sealed, so a new kind cannot be added without a mapping.
 */
public abstract sealed class ApplicationException extends RuntimeException
        permits NotFoundException, ConflictException, ValidationException, UnprocessableException {

    private final String code;
    private final Map<String, Object> details;

    protected ApplicationException(String code, String message, Map<String, Object> details) {
        super(message);
        this.code = code;
        this.details = Map.copyOf(details);
    }

    public String code() {
        return code;
    }

    public Map<String, Object> details() {
        return details;
    }
}
