package geo_test

import (
	"errors"
	"math"
	"testing"

	"ridedispatch/internal/apperr"
	"ridedispatch/internal/geo"
)

func TestNewLocationAcceptsValidCoordinates(t *testing.T) {
	cases := []struct{ lat, lng float64 }{
		{10.7769, 106.7009},
		{-90, -180},
		{90, 180},
		{0, 0},
	}
	for _, c := range cases {
		loc, err := geo.NewLocation(c.lat, c.lng)
		if err != nil {
			t.Fatalf("NewLocation(%v, %v) unexpected error: %v", c.lat, c.lng, err)
		}
		if loc.Latitude != c.lat || loc.Longitude != c.lng {
			t.Fatalf("NewLocation(%v, %v) = %+v", c.lat, c.lng, loc)
		}
	}
}

func TestNewLocationRejectsInvalidCoordinates(t *testing.T) {
	cases := []struct {
		name     string
		lat, lng float64
	}{
		{"latitude above 90", 90.0001, 0},
		{"latitude below -90", -90.0001, 0},
		{"longitude above 180", 0, 180.0001},
		{"longitude below -180", 0, -180.0001},
		{"latitude NaN", math.NaN(), 0},
		{"longitude infinite", 0, math.Inf(1)},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			_, err := geo.NewLocation(c.lat, c.lng)
			if !errors.Is(err, apperr.ErrValidation) {
				t.Fatalf("want ErrValidation, got %v", err)
			}
		})
	}
}

func TestDistanceKm(t *testing.T) {
	origin := geo.MustLocation(10.7769, 106.7009)

	if d := geo.DistanceKm(origin, origin); d != 0 {
		t.Fatalf("distance to self = %v, want 0", d)
	}

	// One degree of latitude is ~111.195 km on a sphere of radius 6371 km.
	north := geo.MustLocation(11.7769, 106.7009)
	if d := geo.DistanceKm(origin, north); math.Abs(d-111.195) > 0.01 {
		t.Fatalf("one degree of latitude = %v km, want ~111.195", d)
	}

	// The brief's pickup and destination, roughly 9.3 km apart in a straight line.
	airport := geo.MustLocation(10.8231, 106.6297)
	if d := geo.DistanceKm(origin, airport); d < 9.2 || d > 9.5 {
		t.Fatalf("brief example distance = %v km, want ~9.3", d)
	}

	if geo.DistanceKm(origin, airport) != geo.DistanceKm(airport, origin) {
		t.Fatal("distance must be symmetric")
	}
}

func TestOffsetNorthKm(t *testing.T) {
	origin := geo.MustLocation(10.7769, 106.7009)
	for _, km := range []float64{0.8, 1.5, 2, 3.2} {
		p := geo.OffsetNorthKm(origin, km)
		if d := geo.DistanceKm(origin, p); math.Abs(d-km) > 1e-6 {
			t.Fatalf("OffsetNorthKm(%v) is %v km away", km, d)
		}
	}
}
