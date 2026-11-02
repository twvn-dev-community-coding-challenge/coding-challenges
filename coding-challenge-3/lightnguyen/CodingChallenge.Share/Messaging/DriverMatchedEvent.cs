namespace CodingChallenge.Share.Messaging;

public sealed record DriverMatchedEvent(Guid RideId, Guid? DriverId);
