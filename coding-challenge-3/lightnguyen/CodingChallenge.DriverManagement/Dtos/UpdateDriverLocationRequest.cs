using System.ComponentModel.DataAnnotations;

namespace CodingChallenge.DriverManagement.Dtos;

public class UpdateDriverLocationRequest
{
    [Range(-90, 90)]
    public required float Latitude { get; set; }

    [Range(-180, 180)]
    public required float Longitude { get; set; }
}
