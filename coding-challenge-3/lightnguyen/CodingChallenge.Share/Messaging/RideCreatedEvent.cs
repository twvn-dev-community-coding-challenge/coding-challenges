namespace CodingChallenge.Share.Messaging;

public sealed record RideCreatedEvent(
    Guid RideId,
    float PickupLatitude,
    float PickupLongitude,
    bool IsCancelled = false);
