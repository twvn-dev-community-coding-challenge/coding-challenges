using CodingChallenge.Share.Models;
using Microsoft.EntityFrameworkCore;

namespace CodingChallenge.Share.Data;

public class BaseDbContext : DbContext
{
    public BaseDbContext(
        DbContextOptions options)
        : base(options)
    {
    }

    public override async Task<int> SaveChangesAsync(
        CancellationToken cancellationToken = default)
    {
        UpdateAuditFields();

        return await base.SaveChangesAsync(cancellationToken);
    }

    public override int SaveChanges()
    {
        UpdateAuditFields();

        return base.SaveChanges();
    }

    private void UpdateAuditFields()
    {
        var entries = ChangeTracker
            .Entries<BaseModel>();

        foreach (var entry in entries)
        {
            if (entry.State == EntityState.Added)
            {
                entry.Entity.CreatedDateTime = DateTime.UtcNow;
            }
            else if (entry.State == EntityState.Modified)
            {
                entry.Entity.UpdatedDateTime = DateTime.UtcNow;
            }
        }
    }
}