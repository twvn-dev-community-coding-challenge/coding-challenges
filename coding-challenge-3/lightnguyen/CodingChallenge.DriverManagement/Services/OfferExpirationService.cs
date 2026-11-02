using CodingChallenge.Share.Messaging;

namespace CodingChallenge.DriverManagement.Services;

public sealed class OfferExpirationService : BackgroundService
{
    private readonly IServiceScopeFactory _scopeFactory;
    private readonly IEventPublisher _publisher;

    public OfferExpirationService(
        IServiceScopeFactory scopeFactory,
        IEventPublisher publisher)
    {
        _scopeFactory = scopeFactory;
        _publisher = publisher;
    }

    protected override async Task ExecuteAsync(CancellationToken stoppingToken)
    {
        using var timer = new PeriodicTimer(TimeSpan.FromSeconds(1));
        while (await timer.WaitForNextTickAsync(stoppingToken))
        {
            await using var scope = _scopeFactory.CreateAsyncScope();
            var driverService = scope.ServiceProvider.GetRequiredService<IDriverService>();
            var expiredOffers = await driverService.ExpirePendingOffers();
            foreach (var offer in expiredOffers)
            {
                await _publisher.PublishAsync(
                    EventRoutingKeys.RideDriverActionCompleted,
                    new RideDriverActionCompletedEvent(
                        offer.RideId,
                        offer.DriverId,
                        RideDriverAction.Reject),
                    stoppingToken);
            }
        }
    }
}
