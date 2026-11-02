using CodingChallenge.DriverManagement.Data;
using CodingChallenge.DriverManagement.Enums;
using CodingChallenge.DriverManagement.Models;
using CodingChallenge.Share.Models;
using Microsoft.EntityFrameworkCore;
using Microsoft.Extensions.Options;

namespace CodingChallenge.DriverManagement.Services;

public class DriverMatchingService
{
    private const double EarthRadiusKilometers = 6371;
    private readonly ApplicationDbContext _dbContext;
    private readonly double _maxRadiusKilometers;

    public DriverMatchingService(
        ApplicationDbContext dbContext,
        IOptions<DriverMatchingOptions> options)
    {
        _dbContext = dbContext;
        _maxRadiusKilometers = options.Value.MaxRadiusKilometers;

        if (!double.IsFinite(_maxRadiusKilometers) || _maxRadiusKilometers < 0)
        {
            throw new ArgumentOutOfRangeException(
                nameof(options),
                "The maximum driver matching radius must be a finite, non-negative value.");
        }
    }

    public async Task<Guid?> FindNearestAvailableDriverId(Coordinate pickup, Guid? rideId = null)
    {
        ArgumentNullException.ThrowIfNull(pickup);

        List<Guid> excludedDriverIds = rideId.HasValue
            ? await _dbContext.Offers
                .Where(offer =>
                    offer.RideId == rideId.Value &&
                    (offer.Status == OfferStatus.REJECTED || offer.Status == OfferStatus.TIMED_OUT))
                .Select(offer => offer.DriverId)
                .Distinct()
                .ToListAsync()
            : [];

        var candidates = await _dbContext.Drivers
            .AsNoTracking()
            .Where(driver =>
                driver.Status == DriverStatus.AVAILABLE &&
                driver.Coordinate != null &&
                !excludedDriverIds.Contains(driver.Id))
            .ToListAsync();

        return candidates
            .Select(driver => new
            {
                driver.Id,
                DistanceKilometers = CalculateDistanceKilometers(pickup, driver.Coordinate!)
            })
            .Where(candidate => candidate.DistanceKilometers <= _maxRadiusKilometers)
            .OrderBy(candidate => candidate.DistanceKilometers)
            .ThenBy(candidate => candidate.Id.ToString("D"), StringComparer.Ordinal)
            .Select(candidate => (Guid?)candidate.Id)
            .FirstOrDefault();
    }

    public Task<Offer?> FindActiveOfferForRide(Guid rideId) =>
        _dbContext.Offers
            .AsNoTracking()
            .Where(offer =>
                offer.RideId == rideId &&
                (offer.Status == OfferStatus.PENDING || offer.Status == OfferStatus.ACCEPTED))
            .OrderByDescending(offer => offer.CreatedDateTime)
            .FirstOrDefaultAsync();

    private static double CalculateDistanceKilometers(Coordinate first, Coordinate second)
    {
        var firstLatitude = DegreesToRadians(first.Latitude);
        var secondLatitude = DegreesToRadians(second.Latitude);
        var latitudeDifference = secondLatitude - firstLatitude;
        var longitudeDifference = DegreesToRadians(second.Longtitude - first.Longtitude);

        var haversine =
            Math.Pow(Math.Sin(latitudeDifference / 2), 2) +
            Math.Cos(firstLatitude) * Math.Cos(secondLatitude) *
            Math.Pow(Math.Sin(longitudeDifference / 2), 2);

        return EarthRadiusKilometers * 2 *
            Math.Asin(Math.Sqrt(Math.Clamp(haversine, 0, 1)));
    }

    private static double DegreesToRadians(double degrees) => degrees * Math.PI / 180;
}
