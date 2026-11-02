using System.ComponentModel.DataAnnotations;

namespace CodingChallenge.Share.Models;

public class BaseModel
{
    [Key]
    public Guid Id { get; set; } = Guid.CreateVersion7();
    public bool IsDeleted { get; set; }
    public DateTime CreatedDateTime { get; set; }
    public DateTime UpdatedDateTime { get; set; }
}