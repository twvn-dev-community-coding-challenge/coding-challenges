// Package driver is the Driver module: the Driver aggregate, its state
// machine, location and storage.
package driver

import (
	"fmt"
	"time"

	"ridedispatch/internal/apperr"
	"ridedispatch/internal/geo"
)

type Status string

const (
	Offline   Status = "OFFLINE"
	Available Status = "AVAILABLE"
	Offered   Status = "OFFERED"
	Busy      Status = "BUSY"
)

var AllStatuses = []Status{Offline, Available, Offered, Busy}

// transitions is the driver lifecycle. Every status change goes through
// moveTo, which refuses anything not listed here.
var transitions = map[Status][]Status{
	Offline:   {Available},
	Available: {Offline, Offered},
	Offered:   {Available, Busy},
	Busy:      {Available},
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

// Driver is the aggregate. Fields are unexported so that other packages can
// only change a driver through its methods. currentRideID is the ride whose
// offer or trip the driver is holding while OFFERED or BUSY.
type Driver struct {
	id                string
	status            Status
	location          geo.Location
	locationUpdatedAt time.Time
	currentRideID     string
	version           int
}

// New creates an OFFLINE driver at loc.
func New(id string, loc geo.Location, now time.Time) Driver {
	return Driver{id: id, status: Offline, location: loc, locationUpdatedAt: now, version: 1}
}

func (d Driver) ID() string                   { return d.id }
func (d Driver) Status() Status               { return d.status }
func (d Driver) Location() geo.Location       { return d.location }
func (d Driver) LocationUpdatedAt() time.Time { return d.locationUpdatedAt }
func (d Driver) CurrentRideID() string        { return d.currentRideID }
func (d Driver) Version() int                 { return d.version }

// SetAvailability is the availability API: it only switches between OFFLINE
// and AVAILABLE. Drivers holding an offer or a trip cannot use it.
func (d *Driver) SetAvailability(target Status, now time.Time) (bool, error) {
	if target != Available && target != Offline {
		return false, fmt.Errorf("%w: availability must be %s or %s, got %q", apperr.ErrValidation, Available, Offline, target)
	}
	if d.status == target {
		return false, nil
	}
	if d.status == Offered || d.status == Busy {
		return false, fmt.Errorf("%w: driver %s is %s and cannot switch to %s", apperr.ErrInvalidTransition, d.id, d.status, target)
	}
	return true, d.moveTo(target, now)
}

// UpdateLocation stores the latest known location. It is allowed in every
// status; matching only looks at AVAILABLE drivers anyway.
func (d *Driver) UpdateLocation(loc geo.Location, now time.Time) {
	d.location = loc
	d.locationUpdatedAt = now
	d.version++
}

// Claim reserves an AVAILABLE driver for an offer on rideID.
func (d *Driver) Claim(rideID string, now time.Time) (bool, error) {
	switch {
	case d.status == Offered && d.currentRideID == rideID:
		return false, nil
	case d.status != Available:
		return false, d.conflict("claim for ride %s", rideID)
	}
	if err := d.moveTo(Offered, now); err != nil {
		return false, err
	}
	d.currentRideID = rideID
	return true, nil
}

// MarkBusy turns the offer on rideID into an accepted trip.
func (d *Driver) MarkBusy(rideID string, now time.Time) (bool, error) {
	switch {
	case d.status == Busy && d.currentRideID == rideID:
		return false, nil
	case d.status != Offered || d.currentRideID != rideID:
		return false, d.conflict("mark busy for ride %s", rideID)
	}
	return true, d.moveTo(Busy, now)
}

// Release frees a driver held by rideID. It is conditional on the ride id: a
// late compensation or timeout for another ride is ignored.
func (d *Driver) Release(rideID string, now time.Time) (bool, error) {
	if (d.status != Offered && d.status != Busy) || d.currentRideID != rideID {
		return false, nil
	}
	if err := d.moveTo(Available, now); err != nil {
		return false, err
	}
	d.currentRideID = ""
	return true, nil
}

func (d *Driver) moveTo(to Status, _ time.Time) error {
	if !CanTransition(d.status, to) {
		return fmt.Errorf("%w: driver %s cannot go from %s to %s", apperr.ErrInvalidTransition, d.id, d.status, to)
	}
	d.status = to
	d.version++
	return nil
}

func (d *Driver) conflict(action string, args ...any) error {
	held := ""
	if d.currentRideID != "" {
		held = " for ride " + d.currentRideID
	}
	return fmt.Errorf("%w: cannot %s: driver %s is %s%s",
		apperr.ErrConflict, fmt.Sprintf(action, args...), d.id, d.status, held)
}
