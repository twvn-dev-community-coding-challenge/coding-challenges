using CodingChallenge.RideManagement.Data;
using CodingChallenge.RideManagement.Dtos;
using CodingChallenge.RideManagement.Enums;
using CodingChallenge.RideManagement.Services;
using CodingChallenge.Share.Data;
using CodingChallenge.Share.Messaging;
using CodingChallenge.Share.Models;
using Microsoft.EntityFrameworkCore;

namespace CodingChallenge.RideManagement.Tests;

public class RideServiceTest
{
    private DbContextOptions<ApplicationDbContext> _options;

    public RideServiceTest()
    {
        _options = new DbContextOptionsBuilder<ApplicationDbContext>()
            .UseInMemoryDatabase(
                databaseName: Guid.NewGuid().ToString())
            .Options;
    }

    private static RideService CreateRideService(
        ApplicationDbContext dbContext,
        FakeEventPublisher eventPublisher) =>
        new(dbContext, eventPublisher);

    [Fact]
    public async Task CreateRide_WithValidData_ShouldSuccess()
    {
        //Arrage
        var idempotencyKey = Guid.NewGuid();

        var request = new CreateRideRequest
        {
            PassengerId = Guid.CreateVersion7(),
            Pickup = new Coordinate(RandomFloat(-90f, 90f), RandomFloat(-180f, 180f)),
            Destination = new Coordinate(RandomFloat(-90f, 90f), RandomFloat(-180f, 180f))
        };

        using var dbContext = new ApplicationDbContext(_options);
        var eventPublisher = new FakeEventPublisher();
        var service = CreateRideService(dbContext, eventPublisher);

        // Act
        var ride = await service.CreateRide(request, idempotencyKey);

        //Assert
        Assert.NotNull(ride);
        Assert.Equal(request.PassengerId, ride.PassengerId);
        Assert.Equal(request.Pickup, ride.Pickup);
        Assert.Equal(request.Destination, ride.Destination);
        Assert.Equal(RideStatus.REQUESTED, ride.Status);
        Assert.Null(ride.AssignedDriverId);
        var publishedEvent = Assert.IsType<RideCreatedEvent>(Assert.Single(eventPublisher.Messages).Message);
        Assert.Equal(ride.Id, publishedEvent.RideId);
        Assert.Equal(request.Pickup.Latitude, publishedEvent.PickupLatitude);
        Assert.Equal(request.Pickup.Longtitude, publishedEvent.PickupLongitude);
    }

    [Fact]
    public async Task CreateRide_WithSameIdempotencyKeyAndRequest_ShouldReturnOriginalResult()
    {
        //Arrage
        var idempotencyKey = Guid.NewGuid();

        var request = new CreateRideRequest
        {
            PassengerId = Guid.CreateVersion7(),
            Pickup = new Coordinate(RandomFloat(-90f, 90f), RandomFloat(-180f, 180f)),
            Destination = new Coordinate(RandomFloat(-90f, 90f), RandomFloat(-180f, 180f))
        };

        using var dbContext = new ApplicationDbContext(_options);
        var eventPublisher = new FakeEventPublisher();
        var service = CreateRideService(dbContext, eventPublisher);

        // Act
        var ride = await service.CreateRide(request, idempotencyKey);
        var ride2 = await service.CreateRide(request, idempotencyKey);
        var ride3 = await service.CreateRide(request, idempotencyKey);

        //Assert
        Assert.NotNull(ride);
        Assert.NotNull(ride2);
        Assert.NotNull(ride3);
        Assert.Equal(ride.Id, ride2.Id);
        Assert.Equal(ride.Id, ride3.Id);
        Assert.Equivalent(ride.Pickup, ride2.Pickup);
        Assert.Equivalent(ride.Destination, ride2.Destination);
        Assert.Equal(ride.Status, ride2.Status);
        Assert.Equal(ride.PassengerId, ride2.PassengerId);
        Assert.Single(dbContext.Rides);
        Assert.Single(dbContext.IdempotencyKeys);
        Assert.Equal(3, eventPublisher.Messages.Count);
    }

    [Fact]
    public async Task CreateRide_WithSameIdempotencyKeyAndDifferentRequest_ShouldRejectKeyReuse()
    {
        //Arrage
        var idempotencyKey = Guid.NewGuid();

        var request = new CreateRideRequest
        {
            PassengerId = Guid.CreateVersion7(),
            Pickup = new Coordinate(RandomFloat(-90f, 90f), RandomFloat(-180f, 180f)),
            Destination = new Coordinate(RandomFloat(-90f, 90f), RandomFloat(-180f, 180f))
        };

        var request2 = new CreateRideRequest
        {
            PassengerId = Guid.CreateVersion7(),
            Pickup = new Coordinate(RandomFloat(-90f, 90f), RandomFloat(-180f, 180f)),
            Destination = new Coordinate(RandomFloat(-90f, 90f), RandomFloat(-180f, 180f))
        };
        using var dbContext = new ApplicationDbContext(_options);
        var service = CreateRideService(dbContext, new FakeEventPublisher());

        // Act
        var ride = await service.CreateRide(request, idempotencyKey);
        await Assert.ThrowsAsync<InvalidOperationException>(
            () => service.CreateRide(request2, idempotencyKey));
    }

    [Fact]
    public async Task GetRideById_AfterCreatingRide_ShouldReturnRide()
    {
        // Arrange
        var idempotencyKey = Guid.NewGuid();

        var request = new CreateRideRequest
        {
            PassengerId = Guid.CreateVersion7(),
            Pickup = new Coordinate(RandomFloat(-90f, 90f), RandomFloat(-180f, 180f)),
            Destination = new Coordinate(RandomFloat(-90f, 90f), RandomFloat(-180f, 180f))
        };
        using var dbContext = new ApplicationDbContext(_options);
        var service = CreateRideService(dbContext, new FakeEventPublisher());

        // Act
        var createdRide = await service.CreateRide(request, idempotencyKey);
        var retrievedRide = await service.GetRideById(createdRide.Id);

        // Assert
        Assert.NotNull(retrievedRide);
        Assert.Equal(createdRide.Id, retrievedRide!.Id);
        Assert.Equal(createdRide.PassengerId, retrievedRide.PassengerId);
        Assert.Equal(createdRide.Pickup, retrievedRide.Pickup);
        Assert.Equal(createdRide.Destination, retrievedRide.Destination);
        Assert.Equal(createdRide.Status, retrievedRide.Status);
    }

    [Fact]
    public async Task RequestOfferResponse_ShouldOnlyPublishForTheAssignedDriver()
    {
        using var dbContext = new ApplicationDbContext(_options);
        var ride = CreateAssignedRide();
        ride.Status = RideStatus.DRIVER_ASSIGNED;
        dbContext.Rides.Add(ride);
        await dbContext.SaveChangesAsync();
        var publisher = new FakeEventPublisher();
        var service = CreateRideService(dbContext, publisher);

        await Assert.ThrowsAsync<InvalidOperationException>(
            () => service.RequestOfferResponse(ride.Id, Guid.NewGuid(), RideDriverAction.Accept));
        Assert.Empty(publisher.Messages);

        ride.Status = RideStatus.DRIVER_ASSIGNED;
        await dbContext.SaveChangesAsync();
        await service.RequestOfferResponse(ride.Id, ride.AssignedDriverId!.Value, RideDriverAction.Accept);

        var message = Assert.IsType<RideDriverActionEvent>(Assert.Single(publisher.Messages).Message);
        Assert.Equal(ride.Id, message.RideId);
        Assert.Equal(ride.AssignedDriverId, message.DriverId);
        Assert.Equal(RideDriverAction.Accept, message.Action);

        await service.HandleDriverActionCompleted(new RideDriverActionCompletedEvent(
            ride.Id,
            ride.AssignedDriverId!.Value,
            RideDriverAction.Accept));
        Assert.Equal(RideStatus.DRIVER_ACCEPTED, ride.Status);
    }

    [Fact]
    public async Task RejectedOffer_ShouldReturnRideToMatchingAndRequestAnotherDriver()
    {
        using var dbContext = new ApplicationDbContext(_options);
        var ride = CreateAssignedRide();
        ride.Status = RideStatus.DRIVER_ASSIGNED;
        var rejectedDriverId = ride.AssignedDriverId!.Value;
        var nextDriverId = Guid.NewGuid();
        dbContext.Rides.Add(ride);
        await dbContext.SaveChangesAsync();
        var publisher = new FakeEventPublisher();
        var service = CreateRideService(dbContext, publisher);

        await service.HandleDriverActionCompleted(new RideDriverActionCompletedEvent(
            ride.Id,
            rejectedDriverId,
            RideDriverAction.Reject));

        Assert.Equal(RideStatus.MATCHING, ride.Status);
        Assert.Null(ride.AssignedDriverId);
        var message = Assert.IsType<RideCreatedEvent>(Assert.Single(publisher.Messages).Message);
        Assert.Equal(ride.Id, message.RideId);

        await service.HandleDriverMatched(new DriverMatchedEvent(ride.Id, nextDriverId));

        Assert.Equal(RideStatus.DRIVER_ASSIGNED, ride.Status);
        Assert.Equal(nextDriverId, ride.AssignedDriverId);
    }

    [Fact]
    public async Task CancelUnassignedRide_ShouldPublishCancellationForDriverCleanup()
    {
        using var dbContext = new ApplicationDbContext(_options);
        var ride = CreateAssignedRide();
        ride.Status = RideStatus.MATCHING;
        ride.AssignedDriverId = null;
        dbContext.Rides.Add(ride);
        await dbContext.SaveChangesAsync();
        var publisher = new FakeEventPublisher();
        var service = CreateRideService(dbContext, publisher);

        await service.CancelRide(ride.Id);

        Assert.Equal(RideStatus.CANCELLED, ride.Status);
        var message = Assert.IsType<RideCreatedEvent>(Assert.Single(publisher.Messages).Message);
        Assert.True(message.IsCancelled);
    }

    [Fact]
    public async Task CancelAcceptedRide_ShouldReleaseItsAssignedDriver()
    {
        using var dbContext = new ApplicationDbContext(_options);
        var ride = CreateAssignedRide();
        ride.Status = RideStatus.DRIVER_ACCEPTED;
        dbContext.Rides.Add(ride);
        await dbContext.SaveChangesAsync();
        var publisher = new FakeEventPublisher();
        var service = CreateRideService(dbContext, publisher);

        await service.CancelRide(ride.Id);

        var action = Assert.IsType<RideDriverActionEvent>(Assert.Single(publisher.Messages).Message);
        Assert.Equal(RideDriverAction.Cancel, action.Action);
        await service.HandleDriverActionCompleted(new RideDriverActionCompletedEvent(
            ride.Id,
            ride.AssignedDriverId!.Value,
            RideDriverAction.Cancel));

        Assert.Equal(RideStatus.CANCELLED, ride.Status);
        Assert.Null(ride.AssignedDriverId);
    }

    [Fact]
    public async Task RideLifecycle_ShouldPersistMatchingAssignmentAcceptanceAndCompletion()
    {
        var ride = CreateAssignedRide();
        ride.Status = RideStatus.REQUESTED;
        var driverId = ride.AssignedDriverId!.Value;
        ride.AssignedDriverId = null;

        using (var dbContext = new ApplicationDbContext(_options))
        {
            dbContext.Rides.Add(ride);
            await dbContext.SaveChangesAsync();
            var publisher = new FakeEventPublisher();
            var service = CreateRideService(dbContext, publisher);

            await service.HandleDriverMatched(new DriverMatchedEvent(ride.Id, driverId));
            Assert.Equal(RideStatus.DRIVER_ASSIGNED, ride.Status);
            Assert.Equal(driverId, ride.AssignedDriverId);

            await service.RequestOfferResponse(ride.Id, driverId, RideDriverAction.Accept);
            await service.HandleDriverActionCompleted(new RideDriverActionCompletedEvent(
                ride.Id, driverId, RideDriverAction.Accept));
            Assert.Equal(RideStatus.DRIVER_ACCEPTED, ride.Status);

            await service.CompleteRide(ride.Id);
            await service.HandleDriverActionCompleted(new RideDriverActionCompletedEvent(
                ride.Id, driverId, RideDriverAction.Complete));
            Assert.Equal(RideStatus.COMPLETED, ride.Status);
        }

        using var reloadedContext = new ApplicationDbContext(_options);
        var persistedRide = await reloadedContext.Rides.FindAsync(ride.Id);
        Assert.NotNull(persistedRide);
        Assert.Equal(RideStatus.COMPLETED, persistedRide.Status);
        Assert.Equal(driverId, persistedRide.AssignedDriverId);
    }

    [Fact]
    public async Task DriverMatchedWithoutEligibleDriver_ShouldLeaveRideMatchingAndUnassigned()
    {
        using var dbContext = new ApplicationDbContext(_options);
        var ride = CreateAssignedRide();
        ride.Status = RideStatus.REQUESTED;
        ride.AssignedDriverId = null;
        dbContext.Rides.Add(ride);
        await dbContext.SaveChangesAsync();
        var service = CreateRideService(dbContext, new FakeEventPublisher());

        await service.HandleDriverMatched(new DriverMatchedEvent(ride.Id, null));

        Assert.Equal(RideStatus.MATCHING, ride.Status);
        Assert.Null(ride.AssignedDriverId);

        using var reloadedContext = new ApplicationDbContext(_options);
        var persistedRide = await reloadedContext.Rides.FindAsync(ride.Id);
        Assert.NotNull(persistedRide);
        Assert.Equal(RideStatus.MATCHING, persistedRide.Status);
        Assert.Null(persistedRide.AssignedDriverId);
    }

    [Fact]
    public async Task DuplicateAndTerminalTransitions_ShouldNotChangeRideState()
    {
        using var dbContext = new ApplicationDbContext(_options);
        var ride = CreateAssignedRide();
        ride.Status = RideStatus.DRIVER_ACCEPTED;
        dbContext.Rides.Add(ride);
        await dbContext.SaveChangesAsync();
        var publisher = new FakeEventPublisher();
        var service = CreateRideService(dbContext, publisher);
        var driverId = ride.AssignedDriverId!.Value;

        await service.HandleDriverActionCompleted(new RideDriverActionCompletedEvent(
            ride.Id, driverId, RideDriverAction.Accept));
        Assert.Equal(RideStatus.DRIVER_ACCEPTED, ride.Status);

        await service.HandleDriverActionCompleted(new RideDriverActionCompletedEvent(
            ride.Id, driverId, RideDriverAction.Cancel));
        Assert.Equal(RideStatus.CANCELLED, ride.Status);
        await service.HandleDriverActionCompleted(new RideDriverActionCompletedEvent(
            ride.Id, driverId, RideDriverAction.Complete));
        await service.HandleDriverMatched(new DriverMatchedEvent(ride.Id, Guid.NewGuid()));

        Assert.Equal(RideStatus.CANCELLED, ride.Status);
        Assert.Null(ride.AssignedDriverId);
        Assert.Empty(publisher.Messages);
        await Assert.ThrowsAsync<InvalidOperationException>(() => service.CompleteRide(ride.Id));
        await Assert.ThrowsAsync<InvalidOperationException>(() => service.CancelRide(ride.Id));
    }

    [Fact]
    public async Task CompletedRide_ShouldRejectFurtherTransitions()
    {
        using var dbContext = new ApplicationDbContext(_options);
        var ride = CreateAssignedRide();
        ride.Status = RideStatus.COMPLETED;
        dbContext.Rides.Add(ride);
        await dbContext.SaveChangesAsync();
        var service = CreateRideService(dbContext, new FakeEventPublisher());
        var driverId = ride.AssignedDriverId!.Value;

        await service.HandleDriverMatched(new DriverMatchedEvent(ride.Id, Guid.NewGuid()));
        await service.HandleDriverActionCompleted(new RideDriverActionCompletedEvent(
            ride.Id, driverId, RideDriverAction.Accept));
        await service.HandleDriverActionCompleted(new RideDriverActionCompletedEvent(
            ride.Id, driverId, RideDriverAction.Cancel));

        Assert.Equal(RideStatus.COMPLETED, ride.Status);
        Assert.Equal(driverId, ride.AssignedDriverId);
        await Assert.ThrowsAsync<InvalidOperationException>(
            () => service.RequestOfferResponse(ride.Id, driverId, RideDriverAction.Accept));
        await Assert.ThrowsAsync<InvalidOperationException>(() => service.CancelRide(ride.Id));
        await Assert.ThrowsAsync<InvalidOperationException>(() => service.CompleteRide(ride.Id));
    }

    [Fact]
    public async Task DuplicateDriverActionEvents_ShouldApplyEachRideTransitionOnlyOnce()
    {
        using var dbContext = new ApplicationDbContext(_options);
        var ride = CreateAssignedRide();
        ride.Status = RideStatus.DRIVER_ASSIGNED;
        var driverId = ride.AssignedDriverId!.Value;
        dbContext.Rides.Add(ride);
        await dbContext.SaveChangesAsync();
        var service = CreateRideService(dbContext, new FakeEventPublisher());
        var acceptedEvent = new RideDriverActionCompletedEvent(
            ride.Id, driverId, RideDriverAction.Accept);
        var completedEvent = new RideDriverActionCompletedEvent(
            ride.Id, driverId, RideDriverAction.Complete);

        await service.HandleDriverActionCompleted(acceptedEvent);
        await service.HandleDriverActionCompleted(acceptedEvent);
        Assert.Equal(RideStatus.DRIVER_ACCEPTED, ride.Status);

        await service.HandleDriverActionCompleted(completedEvent);
        await service.HandleDriverActionCompleted(completedEvent);

        Assert.Equal(RideStatus.COMPLETED, ride.Status);
        using var reloadedContext = new ApplicationDbContext(_options);
        var persistedRide = await reloadedContext.Rides.FindAsync(ride.Id);
        Assert.NotNull(persistedRide);
        Assert.Equal(RideStatus.COMPLETED, persistedRide.Status);
        Assert.Equal(driverId, persistedRide.AssignedDriverId);
    }

    private static CodingChallenge.RideManagement.Models.Ride CreateAssignedRide() =>
        new()
        {
            PassengerId = Guid.NewGuid(),
            Pickup = new Coordinate(0, 0),
            Destination = new Coordinate(1, 1),
            Status = RideStatus.MATCHING,
            AssignedDriverId = Guid.NewGuid()
        };

    private static float RandomFloat(float min, float max)
    {
        return (Random.Shared.NextSingle() * (max - min)) + min;
    }

    private sealed class FakeEventPublisher : IEventPublisher
    {
        public List<(string RoutingKey, object? Message)> Messages { get; } = [];

        public Task PublishAsync<TEvent>(
            string routingKey,
            TEvent message,
            CancellationToken cancellationToken = default)
        {
            Messages.Add((routingKey, message));
            return Task.CompletedTask;
        }
    }
}
