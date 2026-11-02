package dispatch_test

import (
	"context"
	"testing"
	"time"

	"ridedispatch/internal/dispatch"
	"ridedispatch/internal/driver"
	"ridedispatch/internal/geo"
	"ridedispatch/internal/platform/clock"
)

var (
	t0      = time.Date(2026, 10, 31, 9, 0, 0, 0, time.UTC)
	pickup  = geo.MustLocation(10.7769, 106.7009)
	dropoff = geo.MustLocation(10.8231, 106.6297)
)

// fleet registers drivers at the given distances (km, due north of pickup)
// and makes them AVAILABLE.
func fleet(t *testing.T, distances map[string]float64) *driver.Service {
	t.Helper()
	ctx := context.Background()
	svc := driver.NewService(driver.NewMemoryRepository(), clock.NewFake(t0), nil)
	for id, km := range distances {
		if err := svc.Register(ctx, id, geo.OffsetNorthKm(pickup, km)); err != nil {
			t.Fatal(err)
		}
		if _, err := svc.SetAvailability(ctx, id, driver.Available); err != nil {
			t.Fatal(err)
		}
	}
	return svc
}

func ids(cs []dispatch.Candidate) []string {
	out := make([]string, len(cs))
	for i, c := range cs {
		out[i] = c.DriverID
	}
	return out
}

func equal(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func TestMatcherSelectsNearestAndExcludesOutsideRadius(t *testing.T) {
	// The brief's example: A 0.8 km, B 1.5 km, C 3.2 km, radius 2 km.
	drivers := fleet(t, map[string]float64{"C": 3.2, "B": 1.5, "A": 0.8})
	m := dispatch.Matcher{Locator: drivers, RadiusKm: 2}

	got := m.Candidates(context.Background(), pickup, nil)

	if !equal(ids(got), []string{"A", "B"}) {
		t.Fatalf("candidates = %v, want [A B]", ids(got))
	}
	if got[0].DistanceKm < 0.79 || got[0].DistanceKm > 0.81 {
		t.Fatalf("distance to A = %v", got[0].DistanceKm)
	}
}

func TestMatcherRadiusIsInclusive(t *testing.T) {
	drivers := fleet(t, map[string]float64{"D001": 2.0})
	d, _ := drivers.Get(context.Background(), "D001")
	edge := geo.DistanceKm(pickup, d.Location()) // the radius is exactly the driver's distance
	got := dispatch.Matcher{Locator: drivers, RadiusKm: edge}.Candidates(context.Background(), pickup, nil)
	if len(got) != 1 {
		t.Fatalf("driver at the radius edge excluded: %v", got)
	}
}

func TestMatcherRadiusIsConfigurable(t *testing.T) {
	drivers := fleet(t, map[string]float64{"A": 0.8, "C": 3.2})
	got := dispatch.Matcher{Locator: drivers, RadiusKm: 5}.Candidates(context.Background(), pickup, nil)
	if !equal(ids(got), []string{"A", "C"}) {
		t.Fatalf("candidates = %v, want [A C]", ids(got))
	}
}

func TestMatcherDefaultsToTwoKilometres(t *testing.T) {
	drivers := fleet(t, map[string]float64{"A": 1.9, "B": 2.1})
	got := dispatch.Matcher{Locator: drivers}.Candidates(context.Background(), pickup, nil)
	if !equal(ids(got), []string{"A"}) {
		t.Fatalf("candidates = %v, want [A]", ids(got))
	}
}

func TestMatcherExcludesOfflineAndBusyDrivers(t *testing.T) {
	ctx := context.Background()
	drivers := fleet(t, map[string]float64{"OFF": 0.1, "BUSY": 0.2, "OFFERED": 0.3, "FREE": 1.0})
	_, _ = drivers.SetAvailability(ctx, "OFF", driver.Offline)
	_, _ = drivers.Claim(ctx, "BUSY", "R900")
	_, _ = drivers.MarkBusy(ctx, "BUSY", "R900")
	_, _ = drivers.Claim(ctx, "OFFERED", "R901")

	got := dispatch.Matcher{Locator: drivers, RadiusKm: 2}.Candidates(ctx, pickup, nil)
	if !equal(ids(got), []string{"FREE"}) {
		t.Fatalf("candidates = %v, want [FREE]", ids(got))
	}
}

func TestMatcherReturnsNothingWhenNoDriverIsSuitable(t *testing.T) {
	m := dispatch.Matcher{Locator: fleet(t, nil), RadiusKm: 2}
	if got := m.Candidates(context.Background(), pickup, nil); len(got) != 0 {
		t.Fatalf("empty fleet produced %v", got)
	}
	m = dispatch.Matcher{Locator: fleet(t, map[string]float64{"FAR": 5}), RadiusKm: 2}
	if got := m.Candidates(context.Background(), pickup, nil); len(got) != 0 {
		t.Fatalf("far driver produced %v", got)
	}
}

func TestMatcherBreaksTiesByDriverID(t *testing.T) {
	drivers := fleet(t, map[string]float64{"D003": 1, "D001": 1, "D002": 1})
	got := dispatch.Matcher{Locator: drivers, RadiusKm: 2}.Candidates(context.Background(), pickup, nil)
	if !equal(ids(got), []string{"D001", "D002", "D003"}) {
		t.Fatalf("candidates = %v, want lexicographic order", ids(got))
	}
}

func TestMatcherSkipsExcludedDrivers(t *testing.T) {
	drivers := fleet(t, map[string]float64{"A": 0.5, "B": 1.0})
	got := dispatch.Matcher{Locator: drivers, RadiusKm: 2}.Candidates(context.Background(), pickup, map[string]bool{"A": true})
	if !equal(ids(got), []string{"B"}) {
		t.Fatalf("candidates = %v, want [B]", ids(got))
	}
}

func TestMatcherUsesLatestLocation(t *testing.T) {
	ctx := context.Background()
	drivers := fleet(t, map[string]float64{"A": 0.5, "B": 1.0})
	_, _ = drivers.UpdateLocation(ctx, "A", geo.OffsetNorthKm(pickup, 1.8))

	got := dispatch.Matcher{Locator: drivers, RadiusKm: 2}.Candidates(ctx, pickup, nil)
	if !equal(ids(got), []string{"B", "A"}) {
		t.Fatalf("candidates = %v, want [B A] after A moved away", ids(got))
	}
}
