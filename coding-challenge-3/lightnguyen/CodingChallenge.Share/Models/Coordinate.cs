namespace CodingChallenge.Share.Models;

public class Coordinate
{
    private float _latitude;
    private float _longtitude;

    public Coordinate(float latitude, float longtitude)
    {
        Latitude = latitude;
        Longtitude = longtitude;
    }

    public float Latitude
    {
        get => _latitude;
        set
        {
            if (!float.IsFinite(value) || value > 90f || value < -90f)
            {
                throw new ArgumentOutOfRangeException(nameof(Latitude), "Latitude must be a finite value between -90 and 90");
            }

            _latitude = value;
        }
    }
    public float Longtitude
    {
        get => _longtitude;
        set
        {
            if (!float.IsFinite(value) || value > 180f || value < -180f)
            {
                throw new ArgumentOutOfRangeException(nameof(Longtitude), "Longtitude must be a finite value between -180 and 180");
            }

            _longtitude = value;
        }
    }
}