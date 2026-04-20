package org.example.smsgateway.application.handlers;

import org.example.smsgateway.domain.model.command.Command;
import org.example.smsgateway.domain.model.common.Result;

public interface CommandHandler<CommandType extends Command> {
    Result handle(CommandType command);
}
