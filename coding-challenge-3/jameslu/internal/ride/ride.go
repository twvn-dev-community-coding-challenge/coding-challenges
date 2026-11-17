// Package ride is the Ride module: the Ride aggregate, its state machine,
// storage and the ride-creation use case.
package ride

import (
	"fmt"
	"time"

	"ridedispatch/internal/apperr"
	"ridedispatch/internal/geo"
)

type Status string

const (
	Requested      Status = "REQUESTED"
	Matching       Status = "MATCHING"
	DriverAssigned Status = "DRIVER_ASSIGNED"
	DriverAccepted Status = "DRIVER_ACCEPTED"
	Cancelled      Status = "CANCELLED"
	Completed      Status = "COMPLETED"
)

var AllStatuses = []Status{Requested, Matching, DriverAssigned, DriverAccepted, Cancelled, Completed}

// transitions is the ride lifecycle. Every status change goes through moveTo,
// which refuses anything not listed here.
var transitions = map[Status][]Status{
	Requested:      {Matching, Cancelled},
	Matching:       {DriverAssigned, Cancelled},
	DriverAssigned: {Matching, DriverAccepted, Cancelled},
	DriverAccepted: {Completed, Cancelled},
}

// CanTransition reports whether the lifecycle allows from -> to.
func CanTransition(from, to Status) bool {
	for _, s := range transitions[from] {
		if s == to {
			return true
		}
	}
	return false
}

func (s Status) Terminal() bool { return s == Cancelled || s == Completed }

// Ride is the aggregate. Fields are unexported so that other packages can only
// change a ride through its transition methods.
//
// Every transition method returns changed=false with a nil error when the
// transition has already happened (same target, same driver). That makes
// duplicate requests and events safe to apply.
type Ride struct {
	id               string
	passengerID      string
	pickup           geo.Location
	destination      geo.Location
	status           Status
	assignedDriverID string
	createdAt        time.Time
	updatedAt        time.Time
	version          int
}

func New(id, passengerID string, pickup, destination geo.Location, now time.Time) Ride {
	return Ride{
		id:          id,
		passengerID: passengerID,
		pickup:      pickup,
		destination: destination,
		status:      Requested,
		createdAt:   now,
		updatedAt:   now,
		version:     1,
	}
}

func (r Ride) ID() string                { return r.id }
func (r Ride) PassengerID() string       { return r.passengerID }
func (r Ride) Pickup() geo.Location      { return r.pickup }
func (r Ride) Destination() geo.Location { return r.destination }
func (r Ride) Status() Status            { return r.status }
func (r Ride) AssignedDriverID() string  { return r.assignedDriverID }
func (r Ride) CreatedAt() time.Time      { return r.createdAt }
func (r Ride) UpdatedAt() time.Time      { return r.updatedAt }
func (r Ride) Version() int              { return r.version }

// StartMatching moves a new ride into MATCHING. A ride already past
// REQUESTED is left alone.
func (r *Ride) StartMatching(now time.Time) (bool, error) {
	if r.status != Requested {
		return false, nil
	}
	return true, r.moveTo(Matching, now)
}

// Assign records a pending offer to driverID.
func (r *Ride) Assign(driverID string, now time.Time) (bool, error) {
	switch {
	case r.status == DriverAssigned && r.assignedDriverID == driverID:
		return false, nil
	case r.status != Matching:
		return false, r.conflict("assign driver %s", driverID)
	}
	if err := r.moveTo(DriverAssigned, now); err != nil {
		return false, err
	}
	r.assignedDriverID = driverID
	return true, nil
}

// ReleaseAssignment returns the ride to MATCHING after driverID rejected the
// offer or let it expire.
func (r *Ride) ReleaseAssignment(driverID string, now time.Time) (bool, error) {
	switch {
	case r.status == Matching:
		return false, nil
	case r.status != DriverAssigned || r.assignedDriverID != driverID:
		return false, r.conflict("release driver %s", driverID)
	}
	if err := r.moveTo(Matching, now); err != nil {
		return false, err
	}
	r.assignedDriverID = ""
	return true, nil
}

// Accept confirms the pending offer held by driverID.
func (r *Ride) Accept(driverID string, now time.Time) (bool, error) {
	switch {
	case r.status == DriverAccepted && r.assignedDriverID == driverID:
		return false, nil
	case (r.status == DriverAssigned || r.status == DriverAccepted) && r.assignedDriverID != driverID:
		return false, r.conflict("accept by driver %s", driverID)
	}
	return r.change(DriverAccepted, now)
}

// Cancel ends the ride. The assigned driver id is kept for audit.
func (r *Ride) Cancel(now time.Time) (bool, error) {
	if r.status == Cancelled {
		return false, nil
	}
	return r.change(Cancelled, now)
}

// Complete finishes an accepted ride.
func (r *Ride) Complete(now time.Time) (bool, error) {
	if r.status == Completed {
		return false, nil
	}
	return r.change(Completed, now)
}

func (r *Ride) change(to Status, now time.Time) (bool, error) {
	if err := r.moveTo(to, now); err != nil {
		return false, err
	}
	return true, nil
}

func (r *Ride) moveTo(to Status, now time.Time) error {
	if !CanTransition(r.status, to) {
		return fmt.Errorf("%w: ride %s cannot go from %s to %s", apperr.ErrInvalidTransition, r.id, r.status, to)
	}
	r.status = to
	r.updatedAt = now
	r.version++
	return nil
}

func (r *Ride) conflict(action string, args ...any) error {
	held := ""
	if r.assignedDriverID != "" {
		held = " held by " + r.assignedDriverID
	}
	return fmt.Errorf("%w: cannot %s: ride %s is %s%s",
		apperr.ErrConflict, fmt.Sprintf(action, args...), r.id, r.status, held)
}
