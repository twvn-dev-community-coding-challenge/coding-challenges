using CodingChallenge.Share.Messaging;
using Microsoft.Extensions.Options;

namespace CodingChallenge.RideManagement.Services;

public sealed class RideDriverActionCompletedConsumer : RabbitMqConsumer<RideDriverActionCompletedEvent>
{
    private readonly IServiceScopeFactory _scopeFactory;

    public RideDriverActionCompletedConsumer(
        IOptions<RabbitMqOptions> options,
        ILogger<RideDriverActionCompletedConsumer> logger,
        IServiceScopeFactory scopeFactory)
        : base(options, logger)
    {
        _scopeFactory = scopeFactory;
    }

    protected override string QueueName => "ride-management.ride-driver-action-completed";
    protected override string RoutingKey => EventRoutingKeys.RideDriverActionCompleted;

    protected override async Task HandleEventAsync(
        RideDriverActionCompletedEvent message,
        CancellationToken cancellationToken)
    {
        await using var scope = _scopeFactory.CreateAsyncScope();
        var rideService = scope.ServiceProvider.GetRequiredService<IRideService>();
        await rideService.HandleDriverActionCompleted(message);
    }
}
