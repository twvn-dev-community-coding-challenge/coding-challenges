package org.example.smsgateway.domain.model.agreement;

public class Provider {
    private String id;

    public Provider(String providerId) {
        this.id = providerId;
    }

    public String getId() {
        return id;
    }

    @Override
    public boolean equals(Object o) {
        if (this == o) return true;
        if (!(o instanceof Provider other)) return false;
        return id.equals(other.id);
    }

    @Override
    public int hashCode() {
        return id.hashCode();
    }
}
