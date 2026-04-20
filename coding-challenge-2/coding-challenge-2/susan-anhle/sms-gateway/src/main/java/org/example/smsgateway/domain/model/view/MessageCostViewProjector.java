package org.example.smsgateway.domain.model.view;

import org.example.smsgateway.domain.model.common.DomainEvent;
import org.example.smsgateway.domain.model.common.EventRepository;
import org.example.smsgateway.domain.model.message.event.NewMessageRequestReceived;
import org.example.smsgateway.domain.model.message.event.RoutePlanCalculated;
import org.example.smsgateway.domain.model.message.event.smsStatus.*;

import java.util.*;

public class MessageCostViewProjector {

    private final EventRepository eventRepository;

    public MessageCostViewProjector(EventRepository eventRepository) {
        this.eventRepository = eventRepository;
    }

    public Map<String, MessageCostView> project() {
        List<DomainEvent> domainEvents = eventRepository.allEvents();
        Map<String, MessageCostView> result = new HashMap<>();
        for (DomainEvent event : domainEvents) {
            switch (event) {
                case NewMessageRequestReceived e -> {
                    MessageCostView value = new MessageCostView();
                    value.messageId = e.messageId();
                    value.country = e.country();
                    result.put(e.messageId(), value);
                }
                case RoutePlanCalculated e -> {
                    MessageCostView messageCostView = result.get(e.messageId());
                    if (messageCostView == null) break;
                    messageCostView.estimatedCost = e.estimatedCost();
                    messageCostView.providerId = e.providerId();
                }
                case SentSuccessfully e -> {
                    MessageCostView messageCostView = result.get(e.messageId());
                    if (messageCostView == null) break;
                    messageCostView.actualCost = e.actualCost();
                }
                default -> {

                }
            }
        }

        return result;
    }

    public List<CostPerCountry> projectCostPerCountry() {
        Collection<MessageCostView> messageCostViews = project().values();
        List<CostPerCountry> result = new ArrayList<>();
        for (MessageCostView messageCostView : messageCostViews) {
            Optional<CostPerCountry> costPerCountry = findCostPerCountry(result, messageCostView.country);
            if (costPerCountry.isEmpty()) {
                result.add(new CostPerCountry(messageCostView.country, messageCostView.estimatedCost, messageCostView.actualCost));
            } else {
                costPerCountry.get().addEstimatedCost(messageCostView.estimatedCost);
                costPerCountry.get().addActualCost(messageCostView.actualCost);
            }
        }
        return result;
    }

    public List<CostPerProvider> projectCostPerProviderId() {
        Collection<MessageCostView> messageCostViews = project().values();
        List<CostPerProvider> result = new ArrayList<>();
        for (MessageCostView messageCostView : messageCostViews) {
            Optional<CostPerProvider> costPerProvider = findCostPerProvider(result, messageCostView.providerId);
            if (costPerProvider.isEmpty()) {
                result.add(new CostPerProvider(messageCostView.providerId, messageCostView.estimatedCost, messageCostView.actualCost));
            } else {
                costPerProvider.get().addEstimatedCost(messageCostView.estimatedCost);
                costPerProvider.get().addActualCost(messageCostView.actualCost);
            }
        }
        return result;
    }

    private static Optional<CostPerProvider> findCostPerProvider(List<CostPerProvider> list, String providerId) {
        return list.stream().filter(cost -> providerId.equalsIgnoreCase(cost.getProviderId())).findFirst();
    }

    private static Optional<CostPerCountry> findCostPerCountry(List<CostPerCountry> list, String country) {
        return list.stream().filter(cost -> country.equalsIgnoreCase(cost.getCountry())).findFirst();
    }
}
