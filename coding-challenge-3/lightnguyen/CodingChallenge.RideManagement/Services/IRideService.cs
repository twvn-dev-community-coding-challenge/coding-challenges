using CodingChallenge.RideManagement.Dtos;
using CodingChallenge.RideManagement.Models;
using CodingChallenge.Share.Messaging;

namespace CodingChallenge.RideManagement.Services;

public interface IRideService
{
    Task<Ride> CreateRide(CreateRideRequest request, Guid idempotencyKey);
    Task<Ride?> GetRideById(Guid rideId);
    Task HandleDriverMatched(DriverMatchedEvent message);
    Task RequestOfferResponse(Guid rideId, Guid driverId, RideDriverAction action);
    Task CancelRide(Guid rideId);
    Task CompleteRide(Guid rideId);
    Task HandleDriverActionCompleted(RideDriverActionCompletedEvent message);
}
