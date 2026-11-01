package com.ridehailing.common.error;

import java.util.Map;

/** The request is valid, but the current state does not allow it (an illegal transition, an offer not held). */
public final class ConflictException extends ApplicationException {

    public ConflictException(String code, String message, Map<String, Object> details) {
        super(code, message, details);
    }

    public ConflictException(String code, String message) {
        this(code, message, Map.of());
    }
}
