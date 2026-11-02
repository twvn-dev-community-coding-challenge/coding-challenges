using System.ComponentModel.DataAnnotations;

namespace CodingChallenge.RideManagement.Models;

public class IdempotencyKey
{
    [Key]
    public Guid Id { get; set; }
    public required string RequestHash { get; set; }
    public required string Response { get; set; }
}