package ride_test

import (
	"errors"
	"testing"
	"time"

	"ridedispatch/internal/apperr"
	"ridedispatch/internal/geo"
	"ridedispatch/internal/ride"
)

var (
	t0      = time.Date(2026, 10, 31, 9, 0, 0, 0, time.UTC)
	pickup  = geo.MustLocation(10.7769, 106.7009)
	dropoff = geo.MustLocation(10.8231, 106.6297)
)

// rideIn builds a ride and drives it to the wanted status through the public
// transition methods only, so the tests never bypass the state machine.
func rideIn(t *testing.T, status ride.Status, driverID string) ride.Ride {
	t.Helper()
	r := ride.New("R001", "P001", pickup, dropoff, t0)
	steps := map[ride.Status][]func(*ride.Ride) (bool, error){
		ride.Requested:      nil,
		ride.Matching:       {start},
		ride.DriverAssigned: {start, assign(driverID)},
		ride.DriverAccepted: {start, assign(driverID), accept(driverID)},
		ride.Completed:      {start, assign(driverID), accept(driverID), complete},
		ride.Cancelled:      {cancel},
	}
	for _, step := range steps[status] {
		if _, err := step(&r); err != nil {
			t.Fatalf("setup to %s: %v", status, err)
		}
	}
	if r.Status() != status {
		t.Fatalf("setup reached %s, want %s", r.Status(), status)
	}
	return r
}

func start(r *ride.Ride) (bool, error)    { return r.StartMatching(t0) }
func cancel(r *ride.Ride) (bool, error)   { return r.Cancel(t0) }
func complete(r *ride.Ride) (bool, error) { return r.Complete(t0) }
func assign(d string) func(*ride.Ride) (bool, error) {
	return func(r *ride.Ride) (bool, error) { return r.Assign(d, t0) }
}
func accept(d string) func(*ride.Ride) (bool, error) {
	return func(r *ride.Ride) (bool, error) { return r.Accept(d, t0) }
}
func release(d string) func(*ride.Ride) (bool, error) {
	return func(r *ride.Ride) (bool, error) { return r.ReleaseAssignment(d, t0) }
}

func TestNewRideStartsRequested(t *testing.T) {
	r := ride.New("R001", "P001", pickup, dropoff, t0)
	if r.ID() != "R001" || r.PassengerID() != "P001" {
		t.Fatalf("unexpected identity: %s %s", r.ID(), r.PassengerID())
	}
	if r.Status() != ride.Requested {
		t.Fatalf("status = %s, want REQUESTED", r.Status())
	}
	if r.AssignedDriverID() != "" {
		t.Fatalf("assignedDriverId = %q, want empty", r.AssignedDriverID())
	}
	if r.Pickup() != pickup || r.Destination() != dropoff {
		t.Fatal("pickup/destination not stored")
	}
}

func TestRideTransitions(t *testing.T) {
	cases := []struct {
		name        string
		from        ride.Status
		fromDriver  string
		op          func(*ride.Ride) (bool, error)
		wantStatus  ride.Status
		wantDriver  string
		wantChanged bool
		wantErr     error
	}{
		// Valid transitions.
		{"start matching", ride.Requested, "", start, ride.Matching, "", true, nil},
		{"assign", ride.Matching, "", assign("D001"), ride.DriverAssigned, "D001", true, nil},
		{"accept", ride.DriverAssigned, "D001", accept("D001"), ride.DriverAccepted, "D001", true, nil},
		{"complete", ride.DriverAccepted, "D001", complete, ride.Completed, "D001", true, nil},
		{"reject clears driver", ride.DriverAssigned, "D001", release("D001"), ride.Matching, "", true, nil},
		{"cancel requested", ride.Requested, "", cancel, ride.Cancelled, "", true, nil},
		{"cancel matching", ride.Matching, "", cancel, ride.Cancelled, "", true, nil},
		{"cancel assigned", ride.DriverAssigned, "D001", cancel, ride.Cancelled, "D001", true, nil},
		{"cancel accepted", ride.DriverAccepted, "D001", cancel, ride.Cancelled, "D001", true, nil},

		// Duplicate transitions are no-ops.
		{"duplicate start", ride.Matching, "", start, ride.Matching, "", false, nil},
		{"start after assignment", ride.DriverAssigned, "D001", start, ride.DriverAssigned, "D001", false, nil},
		{"duplicate assign same driver", ride.DriverAssigned, "D001", assign("D001"), ride.DriverAssigned, "D001", false, nil},
		{"duplicate accept same driver", ride.DriverAccepted, "D001", accept("D001"), ride.DriverAccepted, "D001", false, nil},
		{"duplicate release", ride.Matching, "", release("D001"), ride.Matching, "", false, nil},
		{"duplicate cancel", ride.Cancelled, "", cancel, ride.Cancelled, "", false, nil},
		{"duplicate complete", ride.Completed, "D001", complete, ride.Completed, "D001", false, nil},

		// Conflicts: the ride is held by another driver.
		{"assign while held by other", ride.DriverAssigned, "D001", assign("D002"), ride.DriverAssigned, "D001", false, apperr.ErrConflict},
		{"accept by other driver", ride.DriverAssigned, "D001", accept("D002"), ride.DriverAssigned, "D001", false, apperr.ErrConflict},
		{"release by other driver", ride.DriverAssigned, "D001", release("D002"), ride.DriverAssigned, "D001", false, apperr.ErrConflict},
		{"release after accept", ride.DriverAccepted, "D001", release("D001"), ride.DriverAccepted, "D001", false, apperr.ErrConflict},
		{"assign cancelled ride", ride.Cancelled, "", assign("D001"), ride.Cancelled, "", false, apperr.ErrConflict},
		{"release cancelled ride", ride.Cancelled, "", release("D001"), ride.Cancelled, "", false, apperr.ErrConflict},

		// Invalid transitions.
		{"accept without offer", ride.Matching, "", accept("D001"), ride.Matching, "", false, apperr.ErrInvalidTransition},
		{"accept cancelled ride", ride.Cancelled, "", accept("D001"), ride.Cancelled, "", false, apperr.ErrInvalidTransition},
		{"complete before accept", ride.DriverAssigned, "D001", complete, ride.DriverAssigned, "D001", false, apperr.ErrInvalidTransition},
		{"complete cancelled ride", ride.Cancelled, "", complete, ride.Cancelled, "", false, apperr.ErrInvalidTransition},
		{"cancel completed ride", ride.Completed, "D001", cancel, ride.Completed, "D001", false, apperr.ErrInvalidTransition},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			r := rideIn(t, c.from, c.fromDriver)
			before := r.Version()

			changed, err := c.op(&r)

			if c.wantErr == nil && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if c.wantErr != nil && !errors.Is(err, c.wantErr) {
				t.Fatalf("error = %v, want %v", err, c.wantErr)
			}
			if changed != c.wantChanged {
				t.Fatalf("changed = %v, want %v", changed, c.wantChanged)
			}
			if r.Status() != c.wantStatus {
				t.Fatalf("status = %s, want %s", r.Status(), c.wantStatus)
			}
			if r.AssignedDriverID() != c.wantDriver {
				t.Fatalf("assignedDriverId = %q, want %q", r.AssignedDriverID(), c.wantDriver)
			}
			if changed && r.Version() != before+1 {
				t.Fatalf("version = %d, want %d", r.Version(), before+1)
			}
			if !changed && r.Version() != before {
				t.Fatalf("no-op changed version %d -> %d", before, r.Version())
			}
		})
	}
}

func TestRideTransitionTable(t *testing.T) {
	allowed := map[[2]ride.Status]bool{
		{ride.Requested, ride.Matching}:            true,
		{ride.Matching, ride.DriverAssigned}:       true,
		{ride.DriverAssigned, ride.Matching}:       true,
		{ride.DriverAssigned, ride.DriverAccepted}: true,
		{ride.DriverAccepted, ride.Completed}:      true,
		{ride.Requested, ride.Cancelled}:           true,
		{ride.Matching, ride.Cancelled}:            true,
		{ride.DriverAssigned, ride.Cancelled}:      true,
		{ride.DriverAccepted, ride.Cancelled}:      true,
	}
	for _, from := range ride.AllStatuses {
		for _, to := range ride.AllStatuses {
			if got := ride.CanTransition(from, to); got != allowed[[2]ride.Status{from, to}] {
				t.Errorf("CanTransition(%s, %s) = %v", from, to, got)
			}
		}
	}
	if ride.CanTransition(ride.Completed, ride.DriverAssigned) {
		t.Fatal("COMPLETED -> DRIVER_ASSIGNED must be rejected")
	}
}

func TestTransitionsRecordUpdatedAt(t *testing.T) {
	r := ride.New("R001", "P001", pickup, dropoff, t0)
	later := t0.Add(time.Minute)
	if _, err := r.StartMatching(later); err != nil {
		t.Fatal(err)
	}
	if !r.UpdatedAt().Equal(later) || !r.CreatedAt().Equal(t0) {
		t.Fatalf("createdAt=%v updatedAt=%v", r.CreatedAt(), r.UpdatedAt())
	}
}
