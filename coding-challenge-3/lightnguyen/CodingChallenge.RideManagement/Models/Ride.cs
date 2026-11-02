using CodingChallenge.RideManagement.Enums;
using CodingChallenge.Share.Models;

namespace CodingChallenge.RideManagement.Models;

public class Ride : BaseModel
{
    public Guid PassengerId { get; set; }
    public required Coordinate Pickup { get; set; }
    public required Coordinate Destination { get; set; }
    public RideStatus Status { get; set; }
    public Guid? AssignedDriverId { get; set; }
}