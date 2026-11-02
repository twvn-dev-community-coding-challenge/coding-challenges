using CodingChallenge.DriverManagement.Enums;
using CodingChallenge.Share.Models;

namespace CodingChallenge.DriverManagement.Models;

public class Driver : BaseModel
{
    public required string Name { get; set; }
    public DriverStatus Status { get; set; }

    public Coordinate? Coordinate { get; set; }

    public void GoOffline()
    {
        if (Status != DriverStatus.AVAILABLE)
        {
            throw new InvalidOperationException("Driver must be available to go offline.");
        }

        Status = DriverStatus.OFFLINE;
    }

    public void ReceiveOffer()
    {
        if (Status != DriverStatus.AVAILABLE)
        {
            throw new InvalidOperationException("Driver must be available to receive an offer.");
        }

        Status = DriverStatus.OFFERED;
    }

    public void AcceptOffer()
    {
        if (Status != DriverStatus.OFFERED)
        {
            throw new InvalidOperationException("Driver must have an offer to accept.");
        }

        Status = DriverStatus.BUSY;
    }

    public void RejectOffer()
    {
        if (Status != DriverStatus.OFFERED)
        {
            throw new InvalidOperationException("Driver must have a pending offer to reject.");
        }

        Status = DriverStatus.AVAILABLE;
    }

    public void CancelOffer()
    {
        if (Status != DriverStatus.OFFERED)
        {
            throw new InvalidOperationException("Driver must have a pending offer to cancel.");
        }

        Status = DriverStatus.AVAILABLE;
    }

    public void CompleteRide()
    {
        if (Status != DriverStatus.BUSY)
        {
            throw new InvalidOperationException("Driver must be busy to complete a ride.");
        }

        Status = DriverStatus.AVAILABLE;
    }
}