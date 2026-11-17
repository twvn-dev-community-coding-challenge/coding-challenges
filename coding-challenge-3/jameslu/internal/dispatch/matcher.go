// Package dispatch is the Dispatch module: driver matching, ride offers and
// the saga that coordinates the Ride and Driver modules.
package dispatch

import (
	"context"
	"sort"

	"ridedispatch/internal/driver"
	"ridedispatch/internal/geo"
)

const DefaultRadiusKm = 2.0

// DriverLocator returns AVAILABLE drivers that may be near a point. It may
// over-approximate (e.g. a whole geo cell); the matcher filters exactly.
type DriverLocator interface {
	AvailableNear(ctx context.Context, p geo.Location, radiusKm float64) []driver.Candidate
}

// Matcher implements "nearest available driver within a radius"
// (docs/adr/ADR-001-matching.md).
type Matcher struct {
	Locator  DriverLocator
	RadiusKm float64 // 0 means DefaultRadiusKm
}

type Candidate struct {
	DriverID   string
	DistanceKm float64
}

// Candidates returns every eligible driver within the radius, nearest first,
// ties broken by driver id. The saga tries them in order, so a lost claim on
// the first falls through to the next.
func (m Matcher) Candidates(ctx context.Context, pickup geo.Location, excluded map[string]bool) []Candidate {
	radius := m.RadiusKm
	if radius <= 0 {
		radius = DefaultRadiusKm
	}
	var out []Candidate
	for _, d := range m.Locator.AvailableNear(ctx, pickup, radius) {
		if excluded[d.ID] {
			continue
		}
		if dist := geo.DistanceKm(pickup, d.Location); dist <= radius {
			out = append(out, Candidate{DriverID: d.ID, DistanceKm: dist})
		}
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].DistanceKm != out[j].DistanceKm {
			return out[i].DistanceKm < out[j].DistanceKm
		}
		return out[i].DriverID < out[j].DriverID
	})
	return out
}
