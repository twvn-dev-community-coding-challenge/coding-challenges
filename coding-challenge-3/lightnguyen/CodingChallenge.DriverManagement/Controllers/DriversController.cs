using CodingChallenge.DriverManagement.Dtos;
using CodingChallenge.DriverManagement.Services;
using Microsoft.AspNetCore.Mvc;

namespace CodingChallenge.DriverManagement.Controllers;

[ApiController]
[Route("api/[controller]")]
public class DriversController : ControllerBase
{
    private readonly IDriverService _driverService;

    public DriversController(IDriverService driverService)
    {
        _driverService = driverService;
    }

    [HttpPost]
    public async Task<IActionResult> CreateDriver([FromBody] CreateDriverRequest request)
    {
        var driver = await _driverService.CreateDriver(request.Name);
        return CreatedAtAction(nameof(GetDriver), new { driverId = driver.Id }, driver);
    }

    [HttpGet("{driverId:guid}")]
    public async Task<IActionResult> GetDriver(Guid driverId)
    {
        var driver = await _driverService.GetDriverById(driverId);
        return driver is null ? NotFound() : Ok(driver);
    }

    [HttpGet("{driverId:guid}/offers")]
    public async Task<IActionResult> GetDriverOffers(Guid driverId)
    {
        try
        {
            return Ok(await _driverService.GetOffersForDriver(driverId));
        }
        catch (KeyNotFoundException exception)
        {
            return NotFound(exception.Message);
        }
    }

    [HttpPut("{driverId}/availability")]
    public async Task<IActionResult> UpdateDriverAvailability(Guid driverId, [FromBody] bool isAvailable)
    {
        await _driverService.UpdateDriverAvailability(driverId, isAvailable);
        return Ok();
    }

    [HttpPut("{driverId:guid}/location")]
    public async Task<IActionResult> UpdateDriverLocation(
        Guid driverId,
        [FromBody] UpdateDriverLocationRequest request)
    {
        var updated = await _driverService.UpdateDriverLocation(
            driverId,
            request.Latitude,
            request.Longitude);

        return updated ? NoContent() : NotFound();
    }
}