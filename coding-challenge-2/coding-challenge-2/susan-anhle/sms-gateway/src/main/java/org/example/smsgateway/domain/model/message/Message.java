package org.example.smsgateway.domain.model.message;

import org.example.smsgateway.domain.model.common.AggregateRoot;
import org.example.smsgateway.domain.model.message.event.smsStatus.Queued;
import org.example.smsgateway.domain.model.message.event.smsStatus.SentToProvider;
import org.example.smsgateway.domain.model.message.event.smsStatus.SentSuccessfully;
import org.example.smsgateway.domain.model.message.event.smsStatus.CarrierRejected;
import org.example.smsgateway.domain.model.message.event.smsStatus.SuccessfullySentToCarrier;
import org.example.smsgateway.domain.model.message.event.MessageEvent;
import org.example.smsgateway.domain.model.message.event.NewMessageRequestReceived;
import org.example.smsgateway.domain.model.message.event.RoutePlanCalculated;

import java.math.BigDecimal;
import java.util.List;

public class Message extends AggregateRoot<MessageEvent, String> {

    private State state;

    public Message(String messageId,
                   String message,
                   String country,
                   String phoneNumber) {
        addNewEvent(new NewMessageRequestReceived(messageId, message, country, phoneNumber));
    }

    public Message(List<MessageEvent> messageEvents) {
        for (MessageEvent messageEvent : messageEvents) {
            updateDataBy(messageEvent);
        }
    }

    @Override
    protected void updateDataBy(MessageEvent event) {
        switch (event) {
            case NewMessageRequestReceived e -> {
                this.state = State.NewMessageRequested;
                this.id = e.messageId();
            }
            case RoutePlanCalculated e -> {
                this.state = State.RoutePlanCalculated;
            }
            case SuccessfullySentToCarrier e -> {
                this.state = State.SentToCarrier;
            }
            case SentSuccessfully e -> {
                this.state = State.SentSuccessfully;
            }
            case Queued e -> {
                this.state = State.Queued;
            }
            case SentToProvider e -> {
                this.state = State.SentToProvider;
            }
            case CarrierRejected e -> {
                this.state = State.CarrierRejected;
            }
            default -> throw new IllegalStateException("Unexpected value: " + event);
        }
    }


    public void addRoutePlan(RoutePlan routePlan) {
        if (!State.NewMessageRequested.equals(this.state))
            throw new IllegalStateException("Can only add route plan when is New-message-requested");
        addNewEvent(new RoutePlanCalculated(id, routePlan.providerId(), routePlan.estimatedCost()));
    }


    public void markAsQueued() {
        if (State.Queued.equals(this.state)) return;
        if (!State.SentToProvider.equals(this.state))
            throw new IllegalStateException("Can only mark as queued when is Sent-to-provider");
        addNewEvent(new Queued(id));
    }

    public void markAsSentToCarrier() {
        if (State.SentToCarrier.equals(this.state)) return;
        if (!State.Queued.equals(this.state))
            throw new IllegalStateException("Can only mark as Sent-to-carrier when is Queued");
        addNewEvent(new SuccessfullySentToCarrier(id));
    }

    public void markAsSentToProvider() {
        if (!State.RoutePlanCalculated.equals(this.state))
            throw new IllegalStateException("Can only mark as Sent-to-provider when is route-plan-calculated");
        addNewEvent(new SentToProvider(id));
    }

    public void markAsCarrierRejected() {
        if (!State.Queued.equals(this.state))
            throw new IllegalStateException("Can only mark as Carrier-rejected when is Queued");
        addNewEvent(new CarrierRejected(id));
    }

    public void markAsSentSuccessfully(BigDecimal actualCost) {
        if (State.SentSuccessfully.equals(this.state)) return;
        if (!State.SentToCarrier.equals(this.state))
            throw new IllegalStateException("Can only mark as Sent-successfully when is Sent-to-carrier");
        addNewEvent(new SentSuccessfully(id, actualCost));
    }

    private enum State {
        NewMessageRequested, RoutePlanCalculated, Queued, SentToCarrier, SentSuccessfully, CarrierRejected, SentToProvider
    }
}
