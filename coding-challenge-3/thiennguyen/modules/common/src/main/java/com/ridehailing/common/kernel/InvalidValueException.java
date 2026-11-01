package com.ridehailing.common.kernel;

/**
 * A value object refused an invalid value (for example a latitude outside ±90). {@code code} is the stable error code
 * the API reports; the web layer maps this to {@code 400 Bad Request}.
 */
public class InvalidValueException extends IllegalArgumentException {

    private final String code;

    public InvalidValueException(String code, String message) {
        super(message);
        this.code = code;
    }

    public String code() {
        return code;
    }
}
