using CodingChallenge.Share.Messaging;
using Microsoft.Extensions.Options;

namespace CodingChallenge.DriverManagement.Services;

public sealed class RideDriverActionConsumer : RabbitMqConsumer<RideDriverActionEvent>
{
    private readonly IServiceScopeFactory _scopeFactory;

    public RideDriverActionConsumer(
        IOptions<RabbitMqOptions> options,
        ILogger<RideDriverActionConsumer> logger,
        IServiceScopeFactory scopeFactory)
        : base(options, logger)
    {
        _scopeFactory = scopeFactory;
    }

    protected override string QueueName => "driver-management.ride-driver-action";
    protected override string RoutingKey => EventRoutingKeys.RideDriverAction;

    protected override async Task HandleEventAsync(
        RideDriverActionEvent message,
        CancellationToken cancellationToken)
    {
        await using var scope = _scopeFactory.CreateAsyncScope();
        var driverService = scope.ServiceProvider.GetRequiredService<IDriverService>();
        var publisher = scope.ServiceProvider.GetRequiredService<IEventPublisher>();

        switch (message.Action)
        {
            case RideDriverAction.Accept:
                await driverService.AcceptOffer(message.RideId, message.DriverId);
                break;
            case RideDriverAction.Reject:
                await driverService.RejectOffer(message.RideId, message.DriverId);
                break;
            case RideDriverAction.Cancel:
                await driverService.CancelOffer(message.RideId, message.DriverId);
                break;
            case RideDriverAction.Complete:
                await driverService.CompleteRide(message.RideId, message.DriverId);
                break;
            default:
                throw new ArgumentOutOfRangeException(nameof(message.Action));
        }

        await publisher.PublishAsync(
            EventRoutingKeys.RideDriverActionCompleted,
            new RideDriverActionCompletedEvent(message.RideId, message.DriverId, message.Action),
            cancellationToken);
    }
}
