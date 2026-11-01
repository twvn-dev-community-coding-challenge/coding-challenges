package com.ridehailing.common.error;

import java.util.Map;

/** The request itself is invalid: a missing or malformed value, or a reference to something unknown. */
public final class ValidationException extends ApplicationException {

    public ValidationException(String code, String message) {
        super(code, message, Map.of());
    }
}
