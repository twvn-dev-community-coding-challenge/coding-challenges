using System.Text.Json.Serialization;
using CodingChallenge.RideManagement.Data;
using CodingChallenge.RideManagement.Services;
using CodingChallenge.Share.Messaging;
using Microsoft.EntityFrameworkCore;

var builder = WebApplication.CreateBuilder(args);

// Add services to the container.
// Learn more about configuring OpenAPI at https://aka.ms/aspnet/openapi
builder.Services.AddOpenApi();

builder.Services.AddControllers()
    .AddJsonOptions(options =>
        options.JsonSerializerOptions.Converters.Add(new JsonStringEnumConverter()))
    .ConfigureApiBehaviorOptions(options =>
    {
        // To preserve the default behavior, capture the original delegate to call later.
        var builtInFactory = options.InvalidModelStateResponseFactory;

        options.InvalidModelStateResponseFactory = context =>
        {
            var logger = context.HttpContext.RequestServices
                                .GetRequiredService<ILogger<RideService>>();

            // Extract validation errors from ModelState
            var errors = context.ModelState
                .Where(e => e.Value?.Errors.Count > 0)
                .ToDictionary(
                    kvp => kvp.Key,
                    kvp => kvp.Value?.Errors.Select(e => e.ErrorMessage).ToArray()
                );

            // Log the 400 validation error
            logger.LogWarning("Validation failed for request {Path}. Errors: {@Errors}", 
                context.HttpContext.Request.Path, 
                errors);
                
            return builtInFactory(context);
        };
    });

var databasePath = Path.GetFullPath(
    builder.Configuration["Database:Path"] ??
    Path.Combine(builder.Environment.ContentRootPath, "data", "Ride.db"));
Directory.CreateDirectory(Path.GetDirectoryName(databasePath)!);
var connectionString = $"Data Source={databasePath}";

builder.Services.AddDbContext<ApplicationDbContext>(options =>
    options.UseSqlite(connectionString)
);

builder.Services.AddScoped<IRideService, RideService>();
builder.Services.Configure<RabbitMqOptions>(builder.Configuration.GetSection("RabbitMq"));
builder.Services.AddSingleton<IEventPublisher, RabbitMqEventPublisher>();
builder.Services.AddHostedService<DriverMatchedConsumer>();
builder.Services.AddHostedService<RideDriverActionCompletedConsumer>();

var app = builder.Build();

await using (var scope = app.Services.CreateAsyncScope())
{
    var dbContext = scope.ServiceProvider.GetRequiredService<ApplicationDbContext>();
    await dbContext.Database.EnsureCreatedAsync();
}

// Configure the HTTP request pipeline.
if (app.Environment.IsDevelopment())
{
    app.MapOpenApi();
}

app.MapControllers();
app.UseHttpsRedirection();

app.Run();
