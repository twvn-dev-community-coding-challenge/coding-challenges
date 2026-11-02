using CodingChallenge.Share.Models;

namespace CodingChallenge.RideManagement.Models;

public class Passenger : BaseModel
{
    public required string Name { get; set; }
}