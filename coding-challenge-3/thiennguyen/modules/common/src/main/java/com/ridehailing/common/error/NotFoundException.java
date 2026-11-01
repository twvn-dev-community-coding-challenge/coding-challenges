package com.ridehailing.common.error;

import java.util.Map;

/** The thing the request names does not exist. */
public final class NotFoundException extends ApplicationException {

    public NotFoundException(String code, String message) {
        super(code, message, Map.of());
    }
}
