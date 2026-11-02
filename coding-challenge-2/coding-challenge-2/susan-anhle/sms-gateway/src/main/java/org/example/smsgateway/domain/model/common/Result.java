package org.example.smsgateway.domain.model.common;

public class Result {

    private boolean isSuccess;
    private final Exception error;

    public boolean isSuccess() {
        return isSuccess;
    }

    public Exception getError() {
        return error;
    }

    private Result(boolean isSuccess) {
        this.isSuccess = isSuccess;
        this.error = null;
    }

    private Result(Exception e) {
        this.isSuccess = false;
        error = e;
    }

    public static Result success() {
        return new Result(true);
    }

    public static Result failure(Exception e) {
        return new Result(e);
    }
}
