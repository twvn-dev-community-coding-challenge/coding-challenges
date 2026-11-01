package com.ridehailing.common.events;

public class MalformedEventException extends RuntimeException {

    public MalformedEventException(String message, Throwable cause) {
        super(message, cause);
    }

    public MalformedEventException(String message) {
        super(message);
    }
}
