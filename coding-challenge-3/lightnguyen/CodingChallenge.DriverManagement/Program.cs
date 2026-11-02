using System.Text.Json.Serialization;
using CodingChallenge.DriverManagement.Data;
using CodingChallenge.DriverManagement.Services;
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
                                .GetRequiredService<ILogger<Program>>();

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
    Path.Combine(builder.Environment.ContentRootPath, "data", "Driver.db"));
Directory.CreateDirectory(Path.GetDirectoryName(databasePath)!);
var connectionString = $"Data Source={databasePath}";

builder.Services.AddDbContext<ApplicationDbContext>(options =>
    options.UseSqlite(connectionString)
);
builder.Services.AddScoped<IDriverService, DriverService>();
builder.Services.Configure<RabbitMqOptions>(builder.Configuration.GetSection("RabbitMq"));
builder.Services.Configure<DriverMatchingOptions>(builder.Configuration.GetSection("DriverMatching"));
builder.Services.Configure<RideOfferOptions>(builder.Configuration.GetSection("RideOffer"));
builder.Services.AddSingleton<IEventPublisher, RabbitMqEventPublisher>();
builder.Services.AddScoped<DriverMatchingService>();
builder.Services.AddHostedService<RideCreatedConsumer>();
builder.Services.AddHostedService<RideDriverActionConsumer>();
builder.Services.AddHostedService<OfferExpirationService>();

var app = builder.Build();

await using (var scope = app.Services.CreateAsyncScope())
{
    var dbContext = scope.ServiceProvider.GetRequiredService<ApplicationDbContext>();
    await dbContext.Database.EnsureCreatedAsync();
    await dbContext.Database.ExecuteSqlRawAsync(
        """CREATE UNIQUE INDEX IF NOT EXISTS "IX_Offers_DriverId" ON "Offers" ("DriverId") WHERE "Status" IN (0, 1)""");
}

// Configure the HTTP request pipeline.
if (app.Environment.IsDevelopment())
{
    app.MapOpenApi();
}

app.MapControllers();
app.UseHttpsRedirection();

app.Run();
