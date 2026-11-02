using CodingChallenge.Share.Models;

namespace CodingChallenge.RideManagement.Dtos;

public class CreateRideRequest
{
    public Guid PassengerId { get; set; }
    public required Coordinate Pickup { get; set; }
    public required Coordinate Destination { get; set; }
}