using CodingChallenge.DriverManagement.Data;
using CodingChallenge.DriverManagement.Enums;
using CodingChallenge.DriverManagement.Models;
using CodingChallenge.Share.Models;
using Microsoft.Data.Sqlite;
using Microsoft.EntityFrameworkCore;
using Microsoft.Extensions.Options;

namespace CodingChallenge.DriverManagement.Services;

public class DriverService : IDriverService
{
    private readonly ApplicationDbContext _dbContext;
    private readonly TimeSpan _offerResponseTimeout;

    public DriverService(
        ApplicationDbContext dbContext,
        IOptions<RideOfferOptions>? options = null)
    {
        _dbContext = dbContext;
        var timeoutSeconds = options?.Value.ResponseTimeoutSeconds ?? 30;
        if (timeoutSeconds <= 0)
        {
            throw new ArgumentOutOfRangeException(
                nameof(options),
                "The offer response timeout must be greater than zero.");
        }

        _offerResponseTimeout = TimeSpan.FromSeconds(timeoutSeconds);
    }

    public async Task<Driver> CreateDriver(string name)
    {
        var driver = new Driver
        {
            Name = name,
            Status = DriverStatus.OFFLINE
        };

        await _dbContext.Drivers.AddAsync(driver);
        await _dbContext.SaveChangesAsync();
        return driver;
    }

    public Task<Driver?> GetDriverById(Guid driverId) =>
        _dbContext.Drivers
            .AsNoTracking()
            .FirstOrDefaultAsync(driver => driver.Id == driverId);

    public async Task<IReadOnlyList<Offer>> GetOffersForDriver(Guid driverId)
    {
        if (!await _dbContext.Drivers.AnyAsync(driver => driver.Id == driverId))
        {
            throw new KeyNotFoundException("Driver not found.");
        }

        return await _dbContext.Offers
            .AsNoTracking()
            .Where(offer => offer.DriverId == driverId)
            .OrderByDescending(offer => offer.CreatedDateTime)
            .ToListAsync();
    }

    public async Task UpdateDriverAvailability(Guid driverId, bool isAvailable)
    {
        var driver = await _dbContext.Drivers.FindAsync(driverId);
        if (driver == null)
        {
            throw new Exception("Driver not found.");
        }

        driver.Status = isAvailable ? DriverStatus.AVAILABLE : DriverStatus.OFFLINE;
        await _dbContext.SaveChangesAsync();
    }

    public async Task<bool> UpdateDriverLocation(Guid driverId, float latitude, float longitude)
    {
        var coordinate = new Coordinate(latitude, longitude);
        var driver = await _dbContext.Drivers.FindAsync(driverId);
        if (driver == null)
        {
            return false;
        }

        driver.Coordinate = coordinate;
        await _dbContext.SaveChangesAsync();
        return true;
    }

    public async Task<bool> SendOfferToDriver(Guid driverId, Guid rideId, decimal price)
    {
        var previouslyRejected = await _dbContext.Offers.AnyAsync(offer =>
            offer.RideId == rideId &&
            offer.DriverId == driverId &&
            (offer.Status == OfferStatus.REJECTED || offer.Status == OfferStatus.TIMED_OUT));
        if (previouslyRejected)
        {
            throw new InvalidOperationException("A driver cannot be offered the same ride more than once.");
        }

        var driver = await _dbContext.Drivers.FindAsync(driverId);
        if (driver == null)
        {
            throw new Exception("Driver not found.");
        }

        driver.ReceiveOffer();

        var offer = new Offer
        {
            DriverId = driverId,
            RideId = rideId,
            Price = price,
            Status = OfferStatus.PENDING
        };

        await _dbContext.Offers.AddAsync(offer);
        try
        {
            await _dbContext.SaveChangesAsync();
            return true;
        }
        catch (DbUpdateException exception) when (
            exception.InnerException is SqliteException
            {
                SqliteExtendedErrorCode: 2067
            })
        {
            _dbContext.Entry(driver).State = EntityState.Detached;
            _dbContext.Entry(offer).State = EntityState.Detached;
            return false;
        }
    }

    public async Task AcceptOffer(Guid offerId)
    {
        var offer = await _dbContext.Offers.FindAsync(offerId);
        if (offer == null)
        {
            throw new Exception("Offer not found.");
        }

        var currentOffer = await GetCurrentOffer(offer.RideId, offer.DriverId);
        if (currentOffer.Id != offer.Id)
        {
            throw new InvalidOperationException("Only the current offer can be accepted.");
        }

        await AcceptOffer(offer.RideId, offer.DriverId);
    }

    public async Task RejectOffer(Guid offerId)
    {
        var offer = await _dbContext.Offers.FindAsync(offerId);
        if (offer == null)
        {
            throw new Exception("Offer not found.");
        }

        var currentOffer = await GetCurrentOffer(offer.RideId, offer.DriverId);
        if (currentOffer.Id != offer.Id)
        {
            throw new InvalidOperationException("Only the current offer can be rejected.");
        }

        await RejectOffer(offer.RideId, offer.DriverId);
    }

    public async Task<Offer> AcceptOffer(Guid rideId, Guid driverId)
    {
        var offer = await GetCurrentOffer(rideId, driverId);
        if (offer.Status == OfferStatus.ACCEPTED)
        {
            return offer;
        }

        EnsurePending(offer);
        var driver = await GetDriver(offer.DriverId);
        driver.AcceptOffer();
        offer.Status = OfferStatus.ACCEPTED;
        await _dbContext.SaveChangesAsync();
        return offer;
    }

    public async Task<Offer> RejectOffer(Guid rideId, Guid driverId, bool timedOut = false)
    {
        var offer = await GetCurrentOffer(rideId, driverId);
        var resolvedStatus = timedOut ? OfferStatus.TIMED_OUT : OfferStatus.REJECTED;
        if (offer.Status == resolvedStatus)
        {
            return offer;
        }

        EnsurePending(offer);
        var driver = await GetDriver(offer.DriverId);
        if (driver.Status != DriverStatus.OFFERED)
        {
            throw new InvalidOperationException("The driver does not hold a pending offer.");
        }

        driver.RejectOffer();
        offer.Status = resolvedStatus;
        await _dbContext.SaveChangesAsync();
        return offer;
    }

    public async Task<Offer> CancelOffer(Guid rideId, Guid driverId)
    {
        var offer = await GetCurrentOffer(rideId, driverId);
        if (offer.Status == OfferStatus.CANCELLED)
        {
            return offer;
        }

        var driver = await GetDriver(offer.DriverId);
        if (offer.Status == OfferStatus.PENDING && driver.Status == DriverStatus.OFFERED)
        {
            driver.CancelOffer();
        }
        else if (offer.Status == OfferStatus.ACCEPTED && driver.Status == DriverStatus.BUSY)
        {
            driver.CompleteRide();
        }
        else
        {
            throw new InvalidOperationException("The offer cannot be cancelled in its current state.");
        }

        offer.Status = OfferStatus.CANCELLED;
        await _dbContext.SaveChangesAsync();
        return offer;
    }

    public async Task<Offer> CompleteRide(Guid rideId, Guid driverId)
    {
        var offer = await GetCurrentOffer(rideId, driverId);
        if (offer.Status == OfferStatus.COMPLETED)
        {
            return offer;
        }

        if (offer.Status != OfferStatus.ACCEPTED)
        {
            throw new InvalidOperationException("Only an accepted ride can be completed.");
        }

        var driver = await GetDriver(offer.DriverId);
        if (driver.Status == DriverStatus.BUSY)
        {
            driver.CompleteRide();
        }
        else if (driver.Status != DriverStatus.AVAILABLE)
        {
            throw new InvalidOperationException("The driver is not completing the accepted ride.");
        }

        offer.Status = OfferStatus.COMPLETED;
        await _dbContext.SaveChangesAsync();
        return offer;
    }

    public async Task CancelPendingOffers(Guid rideId)
    {
        var offers = await _dbContext.Offers
            .Where(offer => offer.RideId == rideId && offer.Status == OfferStatus.PENDING)
            .ToListAsync();

        foreach (var offer in offers)
        {
            var driver = await _dbContext.Drivers.FindAsync(offer.DriverId);
            if (driver?.Status == DriverStatus.OFFERED)
            {
                driver.CancelOffer();
            }

            offer.Status = OfferStatus.CANCELLED;
        }

        if (offers.Count > 0)
        {
            await _dbContext.SaveChangesAsync();
        }
    }

    public async Task<IReadOnlyList<Offer>> ExpirePendingOffers()
    {
        var expiryTime = DateTime.UtcNow - _offerResponseTimeout;
        var expiredOffers = await _dbContext.Offers
            .Where(offer =>
                offer.Status == OfferStatus.PENDING &&
                offer.CreatedDateTime <= expiryTime)
            .ToListAsync();

        foreach (var offer in expiredOffers)
        {
            var driver = await _dbContext.Drivers.FindAsync(offer.DriverId);
            if (driver?.Status == DriverStatus.OFFERED)
            {
                driver.RejectOffer();
            }

            offer.Status = OfferStatus.TIMED_OUT;
        }

        if (expiredOffers.Count > 0)
        {
            await _dbContext.SaveChangesAsync();
        }

        return expiredOffers;
    }

    private async Task<Offer> GetCurrentOffer(Guid rideId, Guid driverId)
    {
        var offer = await _dbContext.Offers
            .Where(candidate => candidate.RideId == rideId && candidate.DriverId == driverId)
            .OrderByDescending(candidate => candidate.CreatedDateTime)
            .ThenByDescending(candidate => candidate.Id)
            .FirstOrDefaultAsync();
        return offer ?? throw new InvalidOperationException("The driver has no offer for this ride.");
    }

    private async Task<Driver> GetDriver(Guid driverId)
    {
        return await _dbContext.Drivers.FindAsync(driverId)
            ?? throw new InvalidOperationException("The driver for this offer was not found.");
    }

    private static void EnsurePending(Offer offer)
    {
        if (offer.Status != OfferStatus.PENDING)
        {
            throw new InvalidOperationException("The offer is no longer pending.");
        }
    }
}