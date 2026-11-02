using System.Text.Json;
using Microsoft.Extensions.Options;
using RabbitMQ.Client;

namespace CodingChallenge.Share.Messaging;

public sealed class RabbitMqEventPublisher : IEventPublisher
{
    public const string ExchangeName = "coding-challenge.events";

    private readonly RabbitMqOptions _options;

    public RabbitMqEventPublisher(IOptions<RabbitMqOptions> options)
    {
        _options = options.Value;
    }

    public Task PublishAsync<TEvent>(
        string routingKey,
        TEvent message,
        CancellationToken cancellationToken = default)
    {
        cancellationToken.ThrowIfCancellationRequested();

        var factory = new ConnectionFactory
        {
            HostName = _options.HostName,
            Port = _options.Port,
            UserName = _options.UserName,
            Password = _options.Password
        };

        using var connection = factory.CreateConnection();
        using var channel = connection.CreateModel();
        RabbitMqTopology.Declare(channel);
        channel.ConfirmSelect();

        var properties = channel.CreateBasicProperties();
        properties.Persistent = true;
        properties.ContentType = "application/json";

        var body = JsonSerializer.SerializeToUtf8Bytes(message);
        channel.BasicPublish(ExchangeName, routingKey, properties, body);
        channel.WaitForConfirmsOrDie(TimeSpan.FromSeconds(5));

        return Task.CompletedTask;
    }
}
