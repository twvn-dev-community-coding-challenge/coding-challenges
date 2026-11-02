using System.Text.Json;
using Microsoft.Extensions.Hosting;
using Microsoft.Extensions.Logging;
using Microsoft.Extensions.Options;
using RabbitMQ.Client;
using RabbitMQ.Client.Events;

namespace CodingChallenge.Share.Messaging;

public abstract class RabbitMqConsumer<TEvent> : BackgroundService
{
    private readonly RabbitMqOptions _options;
    private readonly ILogger _logger;

    protected RabbitMqConsumer(
        IOptions<RabbitMqOptions> options,
        ILogger logger)
    {
        _options = options.Value;
        _logger = logger;
    }

    protected abstract string QueueName { get; }
    protected abstract string RoutingKey { get; }

    protected abstract Task HandleEventAsync(TEvent message, CancellationToken cancellationToken);

    protected override async Task ExecuteAsync(CancellationToken stoppingToken)
    {
        while (!stoppingToken.IsCancellationRequested)
        {
            try
            {
                await ConsumeMessagesAsync(stoppingToken);
            }
            catch (OperationCanceledException) when (stoppingToken.IsCancellationRequested)
            {
                break;
            }
            catch (Exception exception)
            {
                _logger.LogError(exception, "RabbitMQ consumer for {QueueName} stopped.", QueueName);
            }

            if (!stoppingToken.IsCancellationRequested)
            {
                try
                {
                    await Task.Delay(TimeSpan.FromSeconds(5), stoppingToken);
                }
                catch (OperationCanceledException) when (stoppingToken.IsCancellationRequested)
                {
                    break;
                }
            }
        }
    }

    private async Task ConsumeMessagesAsync(CancellationToken stoppingToken)
    {
        var factory = new ConnectionFactory
        {
            HostName = _options.HostName,
            Port = _options.Port,
            UserName = _options.UserName,
            Password = _options.Password,
            AutomaticRecoveryEnabled = true,
            DispatchConsumersAsync = true
        };

        using var connection = factory.CreateConnection();
        using var channel = connection.CreateModel();

        RabbitMqTopology.Declare(channel);
        channel.BasicQos(0, prefetchCount: 1, global: false);

        var consumer = new AsyncEventingBasicConsumer(channel);
        consumer.Received += async (_, eventArgs) =>
        {
            try
            {
                var message = JsonSerializer.Deserialize<TEvent>(eventArgs.Body.Span)
                    ?? throw new InvalidOperationException($"Received an empty {typeof(TEvent).Name} message.");
                await HandleEventAsync(message, stoppingToken);
                channel.BasicAck(eventArgs.DeliveryTag, multiple: false);
            }
            catch (OperationCanceledException) when (stoppingToken.IsCancellationRequested)
            {
                channel.BasicNack(eventArgs.DeliveryTag, multiple: false, requeue: true);
            }
            catch (Exception exception)
            {
                _logger.LogError(exception, "Failed to process a message from {QueueName}; dead-lettering it.", QueueName);
                channel.BasicNack(eventArgs.DeliveryTag, multiple: false, requeue: false);
            }
        };

        channel.BasicConsume(QueueName, autoAck: false, consumer);
        await Task.Delay(Timeout.Infinite, stoppingToken);
    }
}
