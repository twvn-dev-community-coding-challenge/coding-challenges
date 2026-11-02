using CodingChallenge.RideManagement.Models;
using CodingChallenge.Share.Data;
using Microsoft.EntityFrameworkCore;

namespace CodingChallenge.RideManagement.Data;

public class ApplicationDbContext : BaseDbContext
{
    public ApplicationDbContext(
        DbContextOptions<ApplicationDbContext> options)
        : base(options)
    {
    }

    public DbSet<Ride> Rides { get; set; }
    public DbSet<Passenger> Passengers { get; set; }
    public DbSet<IdempotencyKey> IdempotencyKeys { get; set; }

    protected override void OnModelCreating(ModelBuilder modelBuilder)
    {
        base.OnModelCreating(modelBuilder);

        modelBuilder.Entity<Ride>(entity =>
        {
            entity.OwnsOne(r => r.Pickup);
            entity.OwnsOne(r => r.Destination);
        });
    }
}
