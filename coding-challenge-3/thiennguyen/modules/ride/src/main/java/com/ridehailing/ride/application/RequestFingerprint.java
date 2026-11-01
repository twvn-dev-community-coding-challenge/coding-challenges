package com.ridehailing.ride.application;

import java.nio.charset.StandardCharsets;
import java.security.MessageDigest;
import java.security.NoSuchAlgorithmException;
import java.util.HexFormat;

/**
 * SHA-256 over a canonical form of the validated request, to detect an Idempotency-Key reused for another body.
 */
public final class RequestFingerprint {

    private RequestFingerprint() {
    }

    public static String of(CreateRideCommand command) {
        String canonical = String.join("|",
                command.passengerId(),
                Double.toString(command.pickup().latitude()),
                Double.toString(command.pickup().longitude()),
                Double.toString(command.destination().latitude()),
                Double.toString(command.destination().longitude()));
        try {
            byte[] digest = MessageDigest.getInstance("SHA-256").digest(canonical.getBytes(StandardCharsets.UTF_8));
            return HexFormat.of().formatHex(digest);
        } catch (NoSuchAlgorithmException exception) {
            throw new IllegalStateException("SHA-256 is unavailable", exception);
        }
    }
}
