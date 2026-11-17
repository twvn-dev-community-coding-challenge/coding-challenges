// Package geo holds the Location value object and distance calculation.
package geo

import (
	"fmt"
	"math"

	"ridedispatch/internal/apperr"
)

const earthRadiusKm = 6371.0

// Location is a validated latitude/longitude pair in decimal degrees.
type Location struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

// NewLocation validates the coordinates: latitude in [-90, 90] and
// longitude in [-180, 180], both finite.
func NewLocation(lat, lng float64) (Location, error) {
	if math.IsNaN(lat) || math.IsInf(lat, 0) || lat < -90 || lat > 90 {
		return Location{}, fmt.Errorf("%w: latitude %v must be between -90 and 90", apperr.ErrValidation, lat)
	}
	if math.IsNaN(lng) || math.IsInf(lng, 0) || lng < -180 || lng > 180 {
		return Location{}, fmt.Errorf("%w: longitude %v must be between -180 and 180", apperr.ErrValidation, lng)
	}
	return Location{Latitude: lat, Longitude: lng}, nil
}

// MustLocation is NewLocation for constants in seed data and tests.
func MustLocation(lat, lng float64) Location {
	loc, err := NewLocation(lat, lng)
	if err != nil {
		panic(err)
	}
	return loc
}

// DistanceKm returns the great-circle (Haversine) distance between a and b.
func DistanceKm(a, b Location) float64 {
	lat1, lat2 := radians(a.Latitude), radians(b.Latitude)
	dLat := lat2 - lat1
	dLng := radians(b.Longitude - a.Longitude)

	h := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(lat1)*math.Cos(lat2)*math.Sin(dLng/2)*math.Sin(dLng/2)
	return 2 * earthRadiusKm * math.Asin(math.Min(1, math.Sqrt(h)))
}

// OffsetNorthKm returns the point km kilometres due north of p. It is used to
// place seed and test drivers at exact distances from a pickup.
func OffsetNorthKm(p Location, km float64) Location {
	return Location{Latitude: p.Latitude + km/earthRadiusKm*180/math.Pi, Longitude: p.Longitude}
}

func radians(deg float64) float64 { return deg * math.Pi / 180 }
