using CodingChallenge.DriverManagement.Data;
using CodingChallenge.DriverManagement.Enums;
using CodingChallenge.DriverManagement.Models;
using CodingChallenge.DriverManagement.Services;
using CodingChallenge.Share.Data;
using CodingChallenge.Share.Models;
using Microsoft.EntityFrameworkCore;
using Microsoft.Extensions.Options;

namespace CodingChallenge.DriverManagement.Tests;

public class DriverServiceTest
{
    private readonly DbContextOptions<ApplicationDbContext> _options;
    public DriverServiceTest()
    {
        _options = new DbContextOptionsBuilder<ApplicationDbContext>()
            .UseInMemoryDatabase(databaseName: "DriverTest")
            .Options;
    }

    [Fact]
    public async Task CreateDriver_ShouldCreateOfflineDriverAndExposeItsOffers()
    {
        using var dbContext = new ApplicationDbContext(CreateIsolatedOptions());
        var service = new DriverService(dbContext);

        var driver = await service.CreateDriver("Postman driver");
        var offers = await service.GetOffersForDriver(driver.Id);

        Assert.Equal("Postman driver", driver.Name);
        Assert.Equal(DriverStatus.OFFLINE, driver.Status);
        Assert.Empty(offers);
        Assert.Equal(driver.Id, (await service.GetDriverById(driver.Id))?.Id);

        var offer = new Offer
        {
            DriverId = driver.Id,
            RideId = Guid.NewGuid(),
            Price = 0m
        };
        dbContext.Offers.Add(offer);
        await dbContext.SaveChangesAsync();

        Assert.Equal(offer.Id, Assert.Single(await service.GetOffersForDriver(driver.Id)).Id);
    }

    [Fact]
    public async Task FindNearestAvailableDriver_ShouldIgnoreUnavailableAndOutOfRadiusDrivers()
    {
        // Arrange
        var nearestId = Guid.Parse("00000000-0000-0000-0000-000000000001");
        var options = CreateIsolatedOptions();
        using var dbContext = new ApplicationDbContext(options);
        dbContext.Drivers.AddRange(
            new Driver
            {
                Id = nearestId,
                Name = "Nearest",
                Status = DriverStatus.AVAILABLE,
                Coordinate = new Coordinate(0, 0.005f)
            },
            new Driver
            {
                Name = "Farther",
                Status = DriverStatus.AVAILABLE,
                Coordinate = new Coordinate(0, 0.01f)
            },
            new Driver
            {
                Name = "Unavailable",
                Status = DriverStatus.BUSY,
                Coordinate = new Coordinate(0, 0.001f)
            },
            new Driver
            {
                Name = "Outside radius",
                Status = DriverStatus.AVAILABLE,
                Coordinate = new Coordinate(0, 0.03f)
            },
            new Driver
            {
                Name = "Missing location",
                Status = DriverStatus.AVAILABLE
            });
        await dbContext.SaveChangesAsync();

        var matchingService = CreateMatchingService(dbContext);

        // Act
        var selectedDriverId = await matchingService.FindNearestAvailableDriverId(new Coordinate(0, 0));

        // Assert
        Assert.Equal(nearestId, selectedDriverId);
    }

    [Fact]
    public async Task FindNearestAvailableDriver_WhenEquidistant_ShouldSelectLexicographicallySmallestId()
    {
        // Arrange
        var smallerId = Guid.Parse("00000000-0000-0000-0000-000000000001");
        var largerId = Guid.Parse("00000000-0000-0000-0000-000000000002");
        var options = CreateIsolatedOptions();
        using var dbContext = new ApplicationDbContext(options);
        dbContext.Drivers.AddRange(
            new Driver
            {
                Id = largerId,
                Name = "Larger id",
                Status = DriverStatus.AVAILABLE,
                Coordinate = new Coordinate(0, 0.005f)
            },
            new Driver
            {
                Id = smallerId,
                Name = "Smaller id",
                Status = DriverStatus.AVAILABLE,
                Coordinate = new Coordinate(0, -0.005f)
            });
        await dbContext.SaveChangesAsync();

        var matchingService = CreateMatchingService(dbContext);

        // Act
        var selectedDriverId = await matchingService
            .FindNearestAvailableDriverId(new Coordinate(0, 0));

        // Assert
        Assert.Equal(smallerId, selectedDriverId);
    }

    [Fact]
    public async Task FindNearestAvailableDriver_WhenNoSuitableDriverExists_ShouldReturnNull()
    {
        // Arrange
        var options = CreateIsolatedOptions();
        using var dbContext = new ApplicationDbContext(options);
        dbContext.Drivers.Add(new Driver
        {
            Name = "Outside radius",
            Status = DriverStatus.AVAILABLE,
            Coordinate = new Coordinate(0, 0.03f)
        });
        await dbContext.SaveChangesAsync();

        var matchingService = CreateMatchingService(dbContext);

        // Act
        var selectedDriverId = await matchingService
            .FindNearestAvailableDriverId(new Coordinate(0, 0));

        // Assert
        Assert.Null(selectedDriverId);
    }

    [Fact]
    public async Task FindNearestAvailableDriver_ShouldHonorConfiguredRadius()
    {
        // Arrange
        var options = CreateIsolatedOptions();
        using var dbContext = new ApplicationDbContext(options);
        var driver = new Driver
        {
            Name = "Within default radius",
            Status = DriverStatus.AVAILABLE,
            Coordinate = new Coordinate(0, 0.015f)
        };
        dbContext.Drivers.Add(driver);
        await dbContext.SaveChangesAsync();

        var defaultRadiusMatchingService = CreateMatchingService(dbContext, 2);
        var smallerRadiusMatchingService = CreateMatchingService(dbContext, 1.5);

        // Act
        var defaultRadiusMatch = await defaultRadiusMatchingService
            .FindNearestAvailableDriverId(new Coordinate(0, 0));
        var smallerRadiusMatch = await smallerRadiusMatchingService
            .FindNearestAvailableDriverId(new Coordinate(0, 0));

        // Assert
        Assert.Equal(driver.Id, defaultRadiusMatch);
        Assert.Null(smallerRadiusMatch);
    }

    private static DbContextOptions<ApplicationDbContext> CreateIsolatedOptions() =>
        new DbContextOptionsBuilder<ApplicationDbContext>()
            .UseInMemoryDatabase(Guid.NewGuid().ToString())
            .Options;

    private static DriverMatchingService CreateMatchingService(
        ApplicationDbContext dbContext,
        double radiusKilometers = 2) =>
        new(
            dbContext,
            Options.Create(new DriverMatchingOptions
            {
                MaxRadiusKilometers = radiusKilometers
            }));

    [Theory]
    [InlineData(true, DriverStatus.AVAILABLE)]
    [InlineData(false, DriverStatus.OFFLINE)]
    public async Task UpdateDriverAvailability_ShouldUpdateDriverStatus(bool isAvailable, DriverStatus expectedStatus)
    {
        // Arrange
        using var dbContext = new ApplicationDbContext(_options);
        var driver = new Driver
        {
            Id = Guid.NewGuid(),
            Name = "John Doe",
            Status = DriverStatus.OFFLINE
        };

        dbContext.Drivers.Add(driver);
        dbContext.SaveChanges();

        var service = new DriverService(dbContext);

        // Act
        await service.UpdateDriverAvailability(driver.Id, isAvailable);

        // Assert
        var updatedDriver = dbContext.Drivers.Find(driver.Id);
        Assert.NotNull(updatedDriver);
        Assert.Equal(expectedStatus, updatedDriver.Status);
    }

    [Fact]
    public async Task UpdateDriverLocation_ShouldStoreLatestLocation()
    {
        // Arrange
        var driverId = Guid.NewGuid();
        using (var dbContext = new ApplicationDbContext(_options))
        {
            dbContext.Drivers.Add(new Driver
            {
                Id = driverId,
                Name = "Location Test",
                Status = DriverStatus.AVAILABLE,
                Coordinate = new Coordinate(1, 2)
            });
            await dbContext.SaveChangesAsync();
        }

        // Act
        bool firstUpdateResult;
        bool secondUpdateResult;
        using (var dbContext = new ApplicationDbContext(_options))
        {
            var service = new DriverService(dbContext);
            firstUpdateResult = await service.UpdateDriverLocation(driverId, 10, 20);
            secondUpdateResult = await service.UpdateDriverLocation(driverId, -30, -40);
        }

        // Assert
        using var verifyContext = new ApplicationDbContext(_options);
        var updatedDriver = await verifyContext.Drivers.FindAsync(driverId);
        Assert.True(firstUpdateResult);
        Assert.True(secondUpdateResult);
        Assert.NotNull(updatedDriver?.Coordinate);
        Assert.Equal(-30, updatedDriver.Coordinate.Latitude);
        Assert.Equal(-40, updatedDriver.Coordinate.Longtitude);
    }

    [Theory]
    [InlineData(-91, 0)]
    [InlineData(91, 0)]
    [InlineData(0, -181)]
    [InlineData(0, 181)]
    public async Task UpdateDriverLocation_ShouldRejectInvalidCoordinates(float latitude, float longitude)
    {
        // Arrange
        using var dbContext = new ApplicationDbContext(_options);
        var service = new DriverService(dbContext);

        // Act
        var exception = await Record.ExceptionAsync(
            () => service.UpdateDriverLocation(Guid.NewGuid(), latitude, longitude));

        // Assert
        Assert.IsType<ArgumentOutOfRangeException>(exception);
    }

    [Fact]
    public async Task UpdateDriverLocation_ShouldRejectNonFiniteCoordinates()
    {
        // Arrange
        using var dbContext = new ApplicationDbContext(_options);
        var service = new DriverService(dbContext);

        // Act
        var latitudeException = await Record.ExceptionAsync(
            () => service.UpdateDriverLocation(Guid.NewGuid(), float.NaN, 0));
        var longitudeException = await Record.ExceptionAsync(
            () => service.UpdateDriverLocation(Guid.NewGuid(), 0, float.PositiveInfinity));

        // Assert
        Assert.IsType<ArgumentOutOfRangeException>(latitudeException);
        Assert.IsType<ArgumentOutOfRangeException>(longitudeException);
    }

    [Fact]
    public async Task UpdateDriverLocation_ShouldReturnFalseForMissingDriver()
    {
        // Arrange
        using var dbContext = new ApplicationDbContext(_options);
        var service = new DriverService(dbContext);

        // Act
        var result = await service.UpdateDriverLocation(Guid.NewGuid(), 0, 0);

        // Assert
        Assert.False(result);
    }

    [Fact]
    public async Task SendOfferToDriver_ShouldCreateOfferAndUpdateDriverStatus()
    {
        // Arrange
        using var dbContext = new ApplicationDbContext(_options);
        var driver = new Driver
        {
            Id = Guid.NewGuid(),
            Name = "Jane Doe",
            Status = DriverStatus.AVAILABLE
        };

        dbContext.Drivers.Add(driver);
        dbContext.SaveChanges();

        var service = new DriverService(dbContext);
        var rideId = Guid.NewGuid();
        decimal price = 100.0m;

        // Act
        await service.SendOfferToDriver(driver.Id, rideId, price);

        // Assert
        var updatedDriver = dbContext.Drivers.Find(driver.Id);
        var offer = dbContext.Offers.FirstOrDefault(o => o.DriverId == driver.Id && o.RideId == rideId);

        Assert.NotNull(updatedDriver);
        Assert.Equal(DriverStatus.OFFERED, updatedDriver.Status);
        Assert.NotNull(offer);
        Assert.Equal(price, offer.Price);
        Assert.Equal(OfferStatus.PENDING, offer.Status);
    }

    [Fact]
    public async Task SendOfferToDriver_WhenDifferentRidesRaceForOneDriver_ShouldCreateOnlyOneOffer()
    {
        var databasePath = Path.Combine(
            Path.GetTempPath(),
            $"driver-concurrency-{Guid.NewGuid():N}.db");
        var options = new DbContextOptionsBuilder<ApplicationDbContext>()
            .UseSqlite($"Data Source={databasePath}")
            .Options;
        var driverId = Guid.NewGuid();
        var firstRideId = Guid.NewGuid();
        var secondRideId = Guid.NewGuid();

        try
        {
            await using (var setupContext = new ApplicationDbContext(options))
            {
                await setupContext.Database.EnsureCreatedAsync();
                setupContext.Drivers.Add(new Driver
                {
                    Id = driverId,
                    Name = "Concurrent driver",
                    Status = DriverStatus.AVAILABLE
                });
                await setupContext.SaveChangesAsync();
            }

            await using (var firstContext = new ApplicationDbContext(options))
            await using (var secondContext = new ApplicationDbContext(options))
            {
                await firstContext.Drivers.FindAsync(driverId);
                await secondContext.Drivers.FindAsync(driverId);
                var firstService = new DriverService(firstContext);
                var secondService = new DriverService(secondContext);

                var results = await Task.WhenAll(
                    firstService.SendOfferToDriver(driverId, firstRideId, 0m),
                    secondService.SendOfferToDriver(driverId, secondRideId, 0m));

                Assert.Single(results, created => created);
                var offers = await firstContext.Offers.AsNoTracking().ToListAsync();
                Assert.Single(offers);
                Assert.Equal(
                    results[0] ? firstRideId : secondRideId,
                    offers[0].RideId);
            }
        }
        finally
        {
            File.Delete(databasePath);
        }
    }

    [Theory]
    [InlineData(DriverStatus.OFFLINE)]
    [InlineData(DriverStatus.BUSY)]
    [InlineData(DriverStatus.OFFERED)]
    public async Task SendOfferToUnavailableDriver_ShouldThrowException(DriverStatus unavailableStatus)
    {
        // Arrange
        using var dbContext = new ApplicationDbContext(_options);
        var driver = new Driver
        {
            Id = Guid.NewGuid(),
            Name = "Bob Smith",
            Status = unavailableStatus
        };

        dbContext.Drivers.Add(driver);
        dbContext.SaveChanges();

        var service = new DriverService(dbContext);
        var rideId = Guid.NewGuid();
        decimal price = 100.0m;

        // Act
        var exception = await Record.ExceptionAsync(
            () => service.SendOfferToDriver(driver.Id, rideId, price));

        // Assert
        Assert.IsType<InvalidOperationException>(exception);
    }

    [Theory]
    [InlineData(true, OfferStatus.ACCEPTED, DriverStatus.BUSY)]
    [InlineData(false, OfferStatus.REJECTED, DriverStatus.AVAILABLE)]
    public async Task AcceptOrRejectOffer_ShouldUpdateOfferAndDriverStatus(bool accept, OfferStatus expectedOfferStatus, DriverStatus expectedDriverStatus)
    {
        // Arrange
        using var dbContext = new ApplicationDbContext(_options);
        var driver = new Driver
        {
            Id = Guid.NewGuid(),
            Name = "Alice Smith",
            Status = DriverStatus.OFFERED
        };

        var offer = new Offer
        {
            Id = Guid.NewGuid(),
            DriverId = driver.Id,
            RideId = Guid.NewGuid(),
            Price = 150.0m,
            Status = OfferStatus.PENDING
        };

        dbContext.Drivers.Add(driver);
        dbContext.Offers.Add(offer);
        dbContext.SaveChanges();

        var service = new DriverService(dbContext);

        // Act
        if (accept)
        {
            await service.AcceptOffer(offer.Id);
        }
        else
        {
            await service.RejectOffer(offer.Id);
        }

        // Assert
        var updatedOffer = dbContext.Offers.Find(offer.Id);
        var updatedDriver = dbContext.Drivers.Find(driver.Id);

        Assert.NotNull(updatedOffer);
        Assert.Equal(expectedOfferStatus, updatedOffer.Status);
        Assert.NotNull(updatedDriver);
        Assert.Equal(expectedDriverStatus, updatedDriver.Status);
    }

    [Fact]
    public async Task AcceptOffer_WhenRepeated_ShouldKeepOfferAndDriverAcceptedOnce()
    {
        using var dbContext = new ApplicationDbContext(CreateIsolatedOptions());
        var driver = new Driver
        {
            Name = "Idempotent acceptance",
            Status = DriverStatus.AVAILABLE
        };
        dbContext.Drivers.Add(driver);
        await dbContext.SaveChangesAsync();

        var rideId = Guid.NewGuid();
        var service = new DriverService(dbContext);
        await service.SendOfferToDriver(driver.Id, rideId, 100m);

        var firstAcceptance = await service.AcceptOffer(rideId, driver.Id);
        var repeatedAcceptance = await service.AcceptOffer(rideId, driver.Id);

        Assert.Equal(firstAcceptance.Id, repeatedAcceptance.Id);
        Assert.Single(dbContext.Offers);
        Assert.Equal(OfferStatus.ACCEPTED, repeatedAcceptance.Status);
        Assert.Equal(DriverStatus.BUSY, driver.Status);
    }

    [Fact]
    public async Task RejectOffer_ShouldExcludeDriverFromTheSameRideMatching()
    {
        // Arrange
        var options = CreateIsolatedOptions();
        using var dbContext = new ApplicationDbContext(options);
        var firstDriver = new Driver
        {
            Name = "First",
            Status = DriverStatus.AVAILABLE,
            Coordinate = new Coordinate(0, 0.001f)
        };
        var secondDriver = new Driver
        {
            Name = "Second",
            Status = DriverStatus.AVAILABLE,
            Coordinate = new Coordinate(0, 0.002f)
        };
        dbContext.Drivers.AddRange(firstDriver, secondDriver);
        await dbContext.SaveChangesAsync();

        var driverService = new DriverService(dbContext);
        var rideId = Guid.NewGuid();

        // Act
        await driverService.SendOfferToDriver(firstDriver.Id, rideId, 0m);
        await driverService.RejectOffer(rideId, firstDriver.Id);

        var matchingService = CreateMatchingService(dbContext);
        var secondDriverId = await matchingService
            .FindNearestAvailableDriverId(new Coordinate(0, 0), rideId);
        await driverService.SendOfferToDriver(secondDriver.Id, rideId, 0m);
        await driverService.RejectOffer(rideId, secondDriver.Id);

        var noEligibleDriverId = await matchingService
            .FindNearestAvailableDriverId(new Coordinate(0, 0), rideId);
        var driverForAnotherRideId = await matchingService
            .FindNearestAvailableDriverId(new Coordinate(0, 0), Guid.NewGuid());
        var duplicateOfferException = await Record.ExceptionAsync(
            () => driverService.SendOfferToDriver(firstDriver.Id, rideId, 0m));

        // Assert
        Assert.Equal(secondDriver.Id, secondDriverId);
        Assert.Null(noEligibleDriverId);
        Assert.Equal(firstDriver.Id, driverForAnotherRideId);
        Assert.IsType<InvalidOperationException>(duplicateOfferException);
        Assert.Equal(DriverStatus.AVAILABLE, firstDriver.Status);
        Assert.Equal(DriverStatus.AVAILABLE, secondDriver.Status);
    }

    [Fact]
    public async Task ExpirePendingOffers_ShouldReleaseDriverAndMarkOfferTimedOut()
    {
        // Arrange
        var options = CreateIsolatedOptions();
        using var dbContext = new ApplicationDbContext(options);
        var driver = new Driver
        {
            Name = "Timeout",
            Status = DriverStatus.AVAILABLE
        };
        dbContext.Drivers.Add(driver);
        await dbContext.SaveChangesAsync();

        var service = new DriverService(
            dbContext,
            Options.Create(new RideOfferOptions { ResponseTimeoutSeconds = 1 }));
        await service.SendOfferToDriver(driver.Id, Guid.NewGuid(), 0m);
        var offer = dbContext.Offers.First(offer => offer.DriverId == driver.Id);
        offer.CreatedDateTime = DateTime.UtcNow.AddSeconds(-2);
        await dbContext.SaveChangesAsync();

        // Act
        var expiredOffers = await service.ExpirePendingOffers();

        // Assert
        Assert.Contains(offer, expiredOffers);
        Assert.Equal(OfferStatus.TIMED_OUT, offer.Status);
        Assert.Equal(DriverStatus.AVAILABLE, driver.Status);
    }

    [Fact]
    public async Task CancelOfferAndCompleteRide_ShouldReleaseTheDriver()
    {
        // Arrange
        var options = CreateIsolatedOptions();
        using var dbContext = new ApplicationDbContext(options);
        var driver = new Driver
        {
            Name = "Lifecycle",
            Status = DriverStatus.AVAILABLE
        };
        dbContext.Drivers.Add(driver);
        await dbContext.SaveChangesAsync();

        var service = new DriverService(dbContext);
        var cancelledRideId = Guid.NewGuid();

        // Act
        await service.SendOfferToDriver(driver.Id, cancelledRideId, 0m);
        await service.CancelOffer(cancelledRideId, driver.Id);
        var acceptedRideId = Guid.NewGuid();
        await service.SendOfferToDriver(driver.Id, acceptedRideId, 0m);
        await service.AcceptOffer(acceptedRideId, driver.Id);
        var statusAfterAcceptingRide = driver.Status;
        await service.CompleteRide(acceptedRideId, driver.Id);

        var cancelledOffer = dbContext.Offers.Single(offer => offer.RideId == cancelledRideId);
        var acceptedOffer = dbContext.Offers.Single(offer => offer.RideId == acceptedRideId);

        // Assert
        Assert.Equal(OfferStatus.CANCELLED, cancelledOffer.Status);
        Assert.Equal(OfferStatus.COMPLETED, acceptedOffer.Status);
        Assert.Equal(DriverStatus.BUSY, statusAfterAcceptingRide);
        Assert.Equal(DriverStatus.AVAILABLE, driver.Status);
    }

    [Fact]
    public async Task CancelPendingOffers_ShouldReleaseEveryDriverOnTheRide()
    {
        // Arrange
        var options = CreateIsolatedOptions();
        using var dbContext = new ApplicationDbContext(options);
        var drivers = new[]
        {
            new Driver { Name = "First", Status = DriverStatus.AVAILABLE },
            new Driver { Name = "Second", Status = DriverStatus.AVAILABLE }
        };
        dbContext.Drivers.AddRange(drivers);
        await dbContext.SaveChangesAsync();

        var rideId = Guid.NewGuid();
        var service = new DriverService(dbContext);
        foreach (var driver in drivers)
        {
            await service.SendOfferToDriver(driver.Id, rideId, 0m);
        }

        // Act
        await service.CancelPendingOffers(rideId);

        // Assert
        Assert.All(drivers, driver => Assert.Equal(DriverStatus.AVAILABLE, driver.Status));
        Assert.All(
            dbContext.Offers.Where(offer => offer.RideId == rideId),
            offer => Assert.Equal(OfferStatus.CANCELLED, offer.Status));
    }
}
