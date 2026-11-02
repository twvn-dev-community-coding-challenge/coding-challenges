using CodingChallenge.DriverManagement.Models;
using CodingChallenge.Share.Data;
using Microsoft.EntityFrameworkCore;

namespace CodingChallenge.DriverManagement.Data;

public class ApplicationDbContext : BaseDbContext
{
    public ApplicationDbContext(
        DbContextOptions<ApplicationDbContext> options)
        : base(options)
    {
    }

    public DbSet<Driver> Drivers { get; set; }
    public DbSet<Offer> Offers { get; set; }

    protected override void OnModelCreating(ModelBuilder modelBuilder)
    {
        base.OnModelCreating(modelBuilder);

        modelBuilder.Entity<Driver>(entity =>
        {
            entity.OwnsOne(r => r.Coordinate);
        });

        modelBuilder.Entity<Offer>()
            .HasIndex(offer => offer.DriverId)
            .IsUnique()
            .HasFilter("\"Status\" IN (0, 1)");
    }
}
