namespace CodingChallenge.Share.Messaging;

public sealed record RideDriverActionCompletedEvent(
    Guid RideId,
    Guid DriverId,
    RideDriverAction Action);
