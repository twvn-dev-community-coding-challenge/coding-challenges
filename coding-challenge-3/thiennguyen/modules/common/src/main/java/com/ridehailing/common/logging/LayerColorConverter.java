package com.ridehailing.common.logging;

import ch.qos.logback.classic.spi.ILoggingEvent;
import ch.qos.logback.core.pattern.CompositeConverter;
import java.util.List;
import org.springframework.boot.ansi.AnsiColor;
import org.springframework.boot.ansi.AnsiOutput;

/**
 * {@code %layerColor(…)} in the console pattern: colours its content by the layer of the class that logged, so a reader
 * can tell the web, messaging and persistence adapters and the application apart at a glance. Like Spring Boot's
 * {@code %clr}, it emits colour only when ANSI output is enabled ({@code spring.output.ansi.enabled}), so piped output,
 * files and {@code docker compose logs} stay plain.
 */
public class LayerColorConverter extends CompositeConverter<ILoggingEvent> {

    private record Layer(String packageMarker, AnsiColor color) {
    }

    private static final List<Layer> LAYERS = List.of(
            new Layer(".adapters.web.", AnsiColor.BLUE),
            new Layer(".adapters.messaging.", AnsiColor.CYAN),
            new Layer(".adapters.", AnsiColor.MAGENTA),
            new Layer(".application.", AnsiColor.BRIGHT_MAGENTA),
            new Layer(".domain.", AnsiColor.BRIGHT_MAGENTA),
            new Layer(".frameworks.", AnsiColor.BRIGHT_BLUE),
            new Layer("com.ridehailing.common.", AnsiColor.BRIGHT_BLUE));

    static AnsiColor colorOf(String loggerName) {
        if (!loggerName.startsWith("com.ridehailing.")) {
            return AnsiColor.BRIGHT_BLACK;
        }
        return LAYERS.stream()
                .filter(layer -> loggerName.contains(layer.packageMarker()))
                .map(Layer::color)
                .findFirst()
                .orElse(AnsiColor.DEFAULT);
    }

    @Override
    protected String transform(ILoggingEvent event, String in) {
        return AnsiOutput.toString(colorOf(event.getLoggerName()), in);
    }
}
