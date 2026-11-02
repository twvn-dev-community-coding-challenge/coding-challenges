package driver_test

import (
	"errors"
	"testing"
	"time"

	"ridedispatch/internal/apperr"
	"ridedispatch/internal/driver"
	"ridedispatch/internal/geo"
)

var (
	t0   = time.Date(2026, 10, 31, 9, 0, 0, 0, time.UTC)
	here = geo.MustLocation(10.7769, 106.7009)
)

type op func(*driver.Driver) (bool, error)

func available(d *driver.Driver) (bool, error) { return d.SetAvailability(driver.Available, t0) }
func offline(d *driver.Driver) (bool, error)   { return d.SetAvailability(driver.Offline, t0) }
func claim(r string) op {
	return func(d *driver.Driver) (bool, error) { return d.Claim(r, t0) }
}
func markBusy(r string) op {
	return func(d *driver.Driver) (bool, error) { return d.MarkBusy(r, t0) }
}
func release(r string) op {
	return func(d *driver.Driver) (bool, error) { return d.Release(r, t0) }
}

// driverIn drives a new driver to the wanted status through the public
// transition methods only.
func driverIn(t *testing.T, status driver.Status, rideID string) driver.Driver {
	t.Helper()
	d := driver.New("D001", here, t0)
	steps := map[driver.Status][]op{
		driver.Offline:   nil,
		driver.Available: {available},
		driver.Offered:   {available, claim(rideID)},
		driver.Busy:      {available, claim(rideID), markBusy(rideID)},
	}
	for _, step := range steps[status] {
		if _, err := step(&d); err != nil {
			t.Fatalf("setup to %s: %v", status, err)
		}
	}
	return d
}

func TestNewDriverStartsOffline(t *testing.T) {
	d := driver.New("D001", here, t0)
	if d.ID() != "D001" || d.Status() != driver.Offline || d.CurrentRideID() != "" {
		t.Fatalf("unexpected new driver: %s %s %q", d.ID(), d.Status(), d.CurrentRideID())
	}
	if d.Location() != here || !d.LocationUpdatedAt().Equal(t0) {
		t.Fatal("location not stored")
	}
}

func TestDriverTransitions(t *testing.T) {
	cases := []struct {
		name        string
		from        driver.Status
		fromRide    string
		op          op
		wantStatus  driver.Status
		wantRide    string
		wantChanged bool
		wantErr     error
	}{
		// Availability API.
		{"go available", driver.Offline, "", available, driver.Available, "", true, nil},
		{"go offline", driver.Available, "", offline, driver.Offline, "", true, nil},
		{"available twice", driver.Available, "", available, driver.Available, "", false, nil},
		{"offline twice", driver.Offline, "", offline, driver.Offline, "", false, nil},
		{"offered cannot go offline", driver.Offered, "R001", offline, driver.Offered, "R001", false, apperr.ErrInvalidTransition},
		{"busy cannot go offline", driver.Busy, "R001", offline, driver.Busy, "R001", false, apperr.ErrInvalidTransition},
		{"offered cannot set available", driver.Offered, "R001", available, driver.Offered, "R001", false, apperr.ErrInvalidTransition},
		{"busy cannot set available", driver.Busy, "R001", available, driver.Busy, "R001", false, apperr.ErrInvalidTransition},

		// Offer lifecycle.
		{"claim", driver.Available, "", claim("R001"), driver.Offered, "R001", true, nil},
		{"claim again same ride", driver.Offered, "R001", claim("R001"), driver.Offered, "R001", false, nil},
		{"claim offered for other ride", driver.Offered, "R001", claim("R002"), driver.Offered, "R001", false, apperr.ErrConflict},
		{"claim busy", driver.Busy, "R001", claim("R002"), driver.Busy, "R001", false, apperr.ErrConflict},
		{"claim offline", driver.Offline, "", claim("R001"), driver.Offline, "", false, apperr.ErrConflict},
		{"accept", driver.Offered, "R001", markBusy("R001"), driver.Busy, "R001", true, nil},
		{"accept twice", driver.Busy, "R001", markBusy("R001"), driver.Busy, "R001", false, nil},
		{"accept for other ride", driver.Offered, "R001", markBusy("R002"), driver.Offered, "R001", false, apperr.ErrConflict},
		{"accept without offer", driver.Available, "", markBusy("R001"), driver.Available, "", false, apperr.ErrConflict},
		{"reject releases offer", driver.Offered, "R001", release("R001"), driver.Available, "", true, nil},
		{"complete releases busy", driver.Busy, "R001", release("R001"), driver.Available, "", true, nil},
		{"release twice", driver.Available, "", release("R001"), driver.Available, "", false, nil},
		{"stale release ignored", driver.Offered, "R002", release("R001"), driver.Offered, "R002", false, nil},
		{"stale release of busy ignored", driver.Busy, "R002", release("R001"), driver.Busy, "R002", false, nil},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			d := driverIn(t, c.from, c.fromRide)
			before := d.Version()

			changed, err := c.op(&d)

			if c.wantErr == nil && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if c.wantErr != nil && !errors.Is(err, c.wantErr) {
				t.Fatalf("error = %v, want %v", err, c.wantErr)
			}
			if changed != c.wantChanged {
				t.Fatalf("changed = %v, want %v", changed, c.wantChanged)
			}
			if d.Status() != c.wantStatus || d.CurrentRideID() != c.wantRide {
				t.Fatalf("got %s/%q, want %s/%q", d.Status(), d.CurrentRideID(), c.wantStatus, c.wantRide)
			}
			if changed != (d.Version() == before+1) {
				t.Fatalf("version %d -> %d with changed=%v", before, d.Version(), changed)
			}
		})
	}
}

func TestSetAvailabilityRejectsOtherTargets(t *testing.T) {
	d := driverIn(t, driver.Available, "")
	for _, target := range []driver.Status{driver.Offered, driver.Busy, "SLEEPING"} {
		if _, err := d.SetAvailability(target, t0); !errors.Is(err, apperr.ErrValidation) {
			t.Fatalf("SetAvailability(%s) error = %v, want ErrValidation", target, err)
		}
	}
}

func TestDriverTransitionTable(t *testing.T) {
	allowed := map[[2]driver.Status]bool{
		{driver.Offline, driver.Available}: true,
		{driver.Available, driver.Offline}: true,
		{driver.Available, driver.Offered}: true,
		{driver.Offered, driver.Available}: true,
		{driver.Offered, driver.Busy}:      true,
		{driver.Busy, driver.Available}:    true,
	}
	for _, from := range driver.AllStatuses {
		for _, to := range driver.AllStatuses {
			if got := driver.CanTransition(from, to); got != allowed[[2]driver.Status{from, to}] {
				t.Errorf("CanTransition(%s, %s) = %v", from, to, got)
			}
		}
	}
}

func TestUpdateLocationInAnyStatus(t *testing.T) {
	later := t0.Add(time.Minute)
	moved := geo.OffsetNorthKm(here, 1)
	for _, s := range driver.AllStatuses {
		d := driverIn(t, s, "R001")
		d.UpdateLocation(moved, later)
		if d.Location() != moved || !d.LocationUpdatedAt().Equal(later) || d.Status() != s {
			t.Fatalf("status %s: location update not applied", s)
		}
	}
}
