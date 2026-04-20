package org.example.smsgateway.domain.model.common;

import java.util.ArrayList;
import java.util.List;

public abstract class AggregateRoot<TEventType extends DomainEvent, TId> {
    protected TId id;

    public TId id() {
        return id;
    }

    private List<TEventType> changes = new ArrayList<>();

    protected void addNewEvent(TEventType event) {
        updateDataBy(event);
        changes.add(event);
    }

    public List<TEventType> getChanges() {
        return changes;
    }

    protected abstract void updateDataBy(TEventType event);


    public void clearChanges() {
        changes.clear();
    }
}
