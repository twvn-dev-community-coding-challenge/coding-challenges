package com.ridehailing.common.error;

import java.util.Map;

/** The request is well formed but contradicts an earlier one, such as an idempotency key reused with another body. */
public final class UnprocessableException extends ApplicationException {

    public UnprocessableException(String code, String message) {
        super(code, message, Map.of());
    }
}
