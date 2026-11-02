using System.Security.Cryptography;
using System.Text;
using System.Text.Json;

namespace CodingChallenge.RideManagement.Helpers;

public static class RequestHelper
{
    public static string Hash<T>(T request)
    {
        var json = JsonSerializer.Serialize(request);

        var bytes = SHA256.HashData(
            Encoding.UTF8.GetBytes(json));

        return Convert.ToHexString(bytes);
    }
}