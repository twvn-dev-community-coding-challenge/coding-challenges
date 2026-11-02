using CodingChallenge.RideManagement.Dtos;
using CodingChallenge.RideManagement.Services;
using CodingChallenge.Share.Messaging;
using Microsoft.AspNetCore.Mvc;

namespace CodingChallenge.RideManagement.Controllers;

[ApiController]
[Route("[controller]")]
public class RidesController : ControllerBase
{
    private readonly IRideService _rideService;

    public RidesController(IRideService rideService)
    {
        _rideService = rideService;
    }

    [HttpGet("{rideId:guid}")]
    public async Task<IActionResult> GetRide(Guid rideId)
    {
        var ride = await _rideService.GetRideById(rideId);
        return ride is null ? NotFound() : Ok(ride);
    }

    [HttpPost]
    public async Task<IActionResult> CreateRide(CreateRideRequest request, [FromHeader(Name = "Idempotency-Key")] Guid idempotencyKey)
    {
        if (idempotencyKey == Guid.Empty)
        {
            return BadRequest("Idempotency key is required.");
        }

        var result = await _rideService.CreateRide(request, idempotencyKey);

        return Ok(result);
    }

    [HttpPost("{rideId:guid}/offers/{driverId:guid}/accept")]
    public Task<IActionResult> AcceptOffer(Guid rideId, Guid driverId) =>
        RequestOfferResponse(rideId, driverId, RideDriverAction.Accept);

    [HttpPost("{rideId:guid}/offers/{driverId:guid}/reject")]
    public Task<IActionResult> RejectOffer(Guid rideId, Guid driverId) =>
        RequestOfferResponse(rideId, driverId, RideDriverAction.Reject);

    [HttpPost("{rideId:guid}/cancel")]
    public async Task<IActionResult> CancelRide(Guid rideId)
    {
        try
        {
            await _rideService.CancelRide(rideId);
            return Accepted();
        }
        catch (KeyNotFoundException exception)
        {
            return NotFound(exception.Message);
        }
        catch (InvalidOperationException exception)
        {
            return Conflict(exception.Message);
        }
    }

    [HttpPost("{rideId:guid}/complete")]
    public async Task<IActionResult> CompleteRide(Guid rideId)
    {
        try
        {
            await _rideService.CompleteRide(rideId);
            return Accepted();
        }
        catch (KeyNotFoundException exception)
        {
            return NotFound(exception.Message);
        }
        catch (InvalidOperationException exception)
        {
            return Conflict(exception.Message);
        }
    }

    private async Task<IActionResult> RequestOfferResponse(
        Guid rideId,
        Guid driverId,
        RideDriverAction action)
    {
        try
        {
            await _rideService.RequestOfferResponse(rideId, driverId, action);
            return Accepted();
        }
        catch (KeyNotFoundException exception)
        {
            return NotFound(exception.Message);
        }
        catch (InvalidOperationException exception)
        {
            return Conflict(exception.Message);
        }
    }
}