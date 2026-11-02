namespace CodingChallenge.Share.Messaging;

public static class EventRoutingKeys
{
    public const string RideCreated = "ride.created";
    public const string DriverMatched = "driver.matched";
    public const string RideDriverAction = "ride.driver-action";
    public const string RideDriverActionCompleted = "ride.driver-action-completed";
}
