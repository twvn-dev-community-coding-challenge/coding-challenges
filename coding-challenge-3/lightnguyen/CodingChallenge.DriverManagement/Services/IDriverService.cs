using CodingChallenge.DriverManagement.Models;

namespace CodingChallenge.DriverManagement.Services;

public interface IDriverService
{
    Task<Driver> CreateDriver(string name);
    Task<Driver?> GetDriverById(Guid driverId);
    Task<IReadOnlyList<Offer>> GetOffersForDriver(Guid driverId);
    Task UpdateDriverAvailability(Guid driverId, bool isAvailable);
    Task<bool> UpdateDriverLocation(Guid driverId, float latitude, float longitude);
    Task<bool> SendOfferToDriver(Guid driverId, Guid rideId, decimal price);
    Task AcceptOffer(Guid offerId);
    Task RejectOffer(Guid offerId);
    Task<Offer> AcceptOffer(Guid rideId, Guid driverId);
    Task<Offer> RejectOffer(Guid rideId, Guid driverId, bool timedOut = false);
    Task<Offer> CancelOffer(Guid rideId, Guid driverId);
    Task<Offer> CompleteRide(Guid rideId, Guid driverId);
    Task CancelPendingOffers(Guid rideId);
    Task<IReadOnlyList<Offer>> ExpirePendingOffers();
}
