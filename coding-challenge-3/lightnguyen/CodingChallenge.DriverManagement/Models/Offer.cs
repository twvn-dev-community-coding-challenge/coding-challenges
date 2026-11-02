using CodingChallenge.DriverManagement.Enums;
using CodingChallenge.Share.Models;

namespace CodingChallenge.DriverManagement.Models;

public class Offer : BaseModel
{
    public required Guid RideId { get; set; }
    public required Guid DriverId { get; set; }
    public required decimal Price { get; set; }
    public OfferStatus Status { get; set; } = OfferStatus.PENDING;
}