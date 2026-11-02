namespace CodingChallenge.Share.Messaging;

public sealed record RideDriverActionEvent(
    Guid RideId,
    Guid DriverId,
    RideDriverAction Action);
