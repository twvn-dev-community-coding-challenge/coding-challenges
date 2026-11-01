package com.ridehailing.acceptance;

import java.util.ArrayList;
import java.util.List;
import java.util.concurrent.Callable;
import java.util.concurrent.CyclicBarrier;
import java.util.concurrent.ExecutorService;
import java.util.concurrent.Executors;
import java.util.concurrent.Future;
import java.util.concurrent.TimeUnit;

/** Starts every call at the same instant on its own virtual thread and returns the results in call order. */
final class ConcurrentCalls {

    private ConcurrentCalls() {
    }

    static <T> List<T> runTogether(List<Callable<T>> calls) throws Exception {
        CyclicBarrier startLine = new CyclicBarrier(calls.size());
        try (ExecutorService threads = Executors.newVirtualThreadPerTaskExecutor()) {
            List<Future<T>> futures = new ArrayList<>();
            for (Callable<T> call : calls) {
                futures.add(threads.submit(() -> {
                    startLine.await(10, TimeUnit.SECONDS);
                    return call.call();
                }));
            }
            List<T> results = new ArrayList<>();
            for (Future<T> future : futures) {
                results.add(future.get(60, TimeUnit.SECONDS));
            }
            return results;
        }
    }
}
