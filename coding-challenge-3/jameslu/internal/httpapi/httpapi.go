// Package httpapi exposes the modules over REST. It only parses, validates
// shapes and maps errors; all behaviour lives in the modules.
package httpapi

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"ridedispatch/internal/apperr"
	"ridedispatch/internal/dispatch"
	"ridedispatch/internal/driver"
	"ridedispatch/internal/geo"
	"ridedispatch/internal/platform/events"
	"ridedispatch/internal/platform/ids"
	"ridedispatch/internal/platform/trace"
	"ridedispatch/internal/ride"
)

type Deps struct {
	Rides    *ride.Service
	Drivers  *driver.Service
	Dispatch *dispatch.Service
	Timeline interface {
		ForRide(rideID string) []events.Event
	}
	Logger *slog.Logger
}

type api struct{ Deps }

func New(d Deps) http.Handler {
	if d.Logger == nil {
		d.Logger = slog.New(slog.DiscardHandler)
	}
	a := &api{d}
	mux := http.NewServeMux()
	mux.HandleFunc("POST /rides", a.createRide)
	mux.HandleFunc("GET /rides/{rideID}", a.getRide)
	mux.HandleFunc("GET /rides/{rideID}/timeline", a.getTimeline)
	mux.HandleFunc("POST /rides/{rideID}/offers/{driverID}/accept", a.acceptOffer)
	mux.HandleFunc("POST /rides/{rideID}/offers/{driverID}/reject", a.rejectOffer)
	mux.HandleFunc("POST /rides/{rideID}/cancel", a.cancelRide)
	mux.HandleFunc("POST /rides/{rideID}/complete", a.completeRide)
	mux.HandleFunc("GET /drivers/{driverID}", a.getDriver)
	mux.HandleFunc("PUT /drivers/{driverID}/availability", a.setAvailability)
	mux.HandleFunc("PUT /drivers/{driverID}/location", a.updateLocation)
	return a.withRequestID(mux)
}

// withRequestID tags the request with X-Request-ID (generated if absent).
// The id becomes the causation id of any event the request publishes.
func (a *api) withRequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.Header.Get("X-Request-ID")
		if id == "" {
			id = ids.New("REQ")
		}
		w.Header().Set("X-Request-ID", id)
		rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		start := time.Now()
		next.ServeHTTP(rec, r.WithContext(trace.WithCausation(r.Context(), id)))
		a.Logger.Info("http.request", "method", r.Method, "path", r.URL.Path,
			"status", rec.status, "duration_ms", time.Since(start).Milliseconds(), "request_id", id)
	})
}

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (s *statusRecorder) WriteHeader(code int) {
	s.status = code
	s.ResponseWriter.WriteHeader(code)
}

// --- request bodies ---

type coordinates struct {
	Latitude  *float64 `json:"latitude"`
	Longitude *float64 `json:"longitude"`
}

func (c *coordinates) location(field string) (geo.Location, error) {
	if c == nil || c.Latitude == nil || c.Longitude == nil {
		return geo.Location{}, fmt.Errorf("%w: %s.latitude and %s.longitude are required", apperr.ErrValidation, field, field)
	}
	loc, err := geo.NewLocation(*c.Latitude, *c.Longitude)
	if err != nil {
		return geo.Location{}, fmt.Errorf("%s: %w", field, err)
	}
	return loc, nil
}

type createRideBody struct {
	PassengerID string       `json:"passengerId"`
	Pickup      *coordinates `json:"pickup"`
	Destination *coordinates `json:"destination"`
}

type availabilityBody struct {
	Status driver.Status `json:"status"`
}

// --- handlers ---

func (a *api) createRide(w http.ResponseWriter, r *http.Request) {
	key := r.Header.Get("Idempotency-Key")
	if key == "" {
		a.fail(w, r, fmt.Errorf("%w: send an Idempotency-Key header", apperr.ErrIdempotencyKeyMissing))
		return
	}
	var body createRideBody
	if !a.decode(w, r, &body) {
		return
	}
	pickup, err := body.Pickup.location("pickup")
	if err != nil {
		a.fail(w, r, err)
		return
	}
	destination, err := body.Destination.location("destination")
	if err != nil {
		a.fail(w, r, err)
		return
	}

	created, replayed, err := a.Rides.Create(r.Context(), key, ride.CreateRequest{
		PassengerID: body.PassengerID, Pickup: pickup, Destination: destination,
	})
	if err != nil {
		a.fail(w, r, err)
		return
	}
	if replayed {
		w.Header().Set("Idempotent-Replayed", "true")
	}
	w.Header().Set("Location", "/rides/"+created.ID())
	writeJSON(w, http.StatusCreated, rideView(created))
}

func (a *api) getRide(w http.ResponseWriter, r *http.Request) {
	got, err := a.Rides.Get(r.Context(), r.PathValue("rideID"))
	a.respondRide(w, r, got, err)
}

func (a *api) getTimeline(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("rideID")
	if _, err := a.Rides.Get(r.Context(), id); err != nil {
		a.fail(w, r, err)
		return
	}
	evs := a.Timeline.ForRide(id)
	if evs == nil {
		evs = []events.Event{}
	}
	writeJSON(w, http.StatusOK, map[string]any{"rideId": id, "events": evs})
}

func (a *api) acceptOffer(w http.ResponseWriter, r *http.Request) {
	got, err := a.Dispatch.Accept(r.Context(), r.PathValue("rideID"), r.PathValue("driverID"))
	a.respondRide(w, r, got, err)
}

func (a *api) rejectOffer(w http.ResponseWriter, r *http.Request) {
	got, err := a.Dispatch.Reject(r.Context(), r.PathValue("rideID"), r.PathValue("driverID"))
	a.respondRide(w, r, got, err)
}

func (a *api) cancelRide(w http.ResponseWriter, r *http.Request) {
	got, err := a.Dispatch.Cancel(r.Context(), r.PathValue("rideID"))
	a.respondRide(w, r, got, err)
}

func (a *api) completeRide(w http.ResponseWriter, r *http.Request) {
	got, err := a.Dispatch.Complete(r.Context(), r.PathValue("rideID"))
	a.respondRide(w, r, got, err)
}

func (a *api) getDriver(w http.ResponseWriter, r *http.Request) {
	d, err := a.Drivers.Get(r.Context(), r.PathValue("driverID"))
	a.respondDriver(w, r, d, err)
}

func (a *api) setAvailability(w http.ResponseWriter, r *http.Request) {
	var body availabilityBody
	if !a.decode(w, r, &body) {
		return
	}
	d, err := a.Drivers.SetAvailability(r.Context(), r.PathValue("driverID"), body.Status)
	a.respondDriver(w, r, d, err)
}

func (a *api) updateLocation(w http.ResponseWriter, r *http.Request) {
	var body coordinates
	if !a.decode(w, r, &body) {
		return
	}
	loc, err := body.location("location")
	if err != nil {
		a.fail(w, r, err)
		return
	}
	d, err := a.Drivers.UpdateLocation(r.Context(), r.PathValue("driverID"), loc)
	a.respondDriver(w, r, d, err)
}

// --- responses ---

type rideJSON struct {
	RideID           string       `json:"rideId"`
	PassengerID      string       `json:"passengerId"`
	Pickup           geo.Location `json:"pickup"`
	Destination      geo.Location `json:"destination"`
	Status           ride.Status  `json:"status"`
	AssignedDriverID *string      `json:"assignedDriverId"`
	CreatedAt        time.Time    `json:"createdAt"`
	UpdatedAt        time.Time    `json:"updatedAt"`
}

func rideView(r ride.Ride) rideJSON {
	return rideJSON{
		RideID:           r.ID(),
		PassengerID:      r.PassengerID(),
		Pickup:           r.Pickup(),
		Destination:      r.Destination(),
		Status:           r.Status(),
		AssignedDriverID: optional(r.AssignedDriverID()),
		CreatedAt:        r.CreatedAt(),
		UpdatedAt:        r.UpdatedAt(),
	}
}

type driverJSON struct {
	DriverID          string        `json:"driverId"`
	Status            driver.Status `json:"status"`
	Location          geo.Location  `json:"location"`
	LocationUpdatedAt time.Time     `json:"locationUpdatedAt"`
	CurrentRideID     *string       `json:"currentRideId"`
}

func driverView(d driver.Driver) driverJSON {
	return driverJSON{
		DriverID:          d.ID(),
		Status:            d.Status(),
		Location:          d.Location(),
		LocationUpdatedAt: d.LocationUpdatedAt(),
		CurrentRideID:     optional(d.CurrentRideID()),
	}
}

func optional(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func (a *api) respondRide(w http.ResponseWriter, r *http.Request, got ride.Ride, err error) {
	if err != nil {
		a.fail(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, rideView(got))
}

func (a *api) respondDriver(w http.ResponseWriter, r *http.Request, d driver.Driver, err error) {
	if err != nil {
		a.fail(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, driverView(d))
}

func (a *api) decode(w http.ResponseWriter, r *http.Request, into any) bool {
	if err := json.NewDecoder(r.Body).Decode(into); err != nil {
		a.fail(w, r, fmt.Errorf("%w: invalid JSON body: %v", apperr.ErrValidation, err))
		return false
	}
	return true
}

// errorMapping is checked in order; the first matching category wins.
var errorMapping = []struct {
	err    error
	status int
	code   string
}{
	{apperr.ErrIdempotencyKeyMissing, http.StatusBadRequest, "IDEMPOTENCY_KEY_MISSING"},
	{apperr.ErrIdempotencyKeyReused, http.StatusUnprocessableEntity, "IDEMPOTENCY_KEY_REUSED"},
	{apperr.ErrIdempotencyInProgress, http.StatusConflict, "IDEMPOTENCY_IN_PROGRESS"},
	{apperr.ErrValidation, http.StatusBadRequest, "VALIDATION_FAILED"},
	{apperr.ErrNotFound, http.StatusNotFound, "NOT_FOUND"},
	{apperr.ErrInvalidTransition, http.StatusConflict, "INVALID_TRANSITION"},
	{apperr.ErrConflict, http.StatusConflict, "CONFLICT"},
}

func (a *api) fail(w http.ResponseWriter, r *http.Request, err error) {
	status, code := http.StatusInternalServerError, "INTERNAL"
	for _, m := range errorMapping {
		if errors.Is(err, m.err) {
			status, code = m.status, m.code
			break
		}
	}
	message := err.Error()
	if status == http.StatusInternalServerError {
		a.Logger.Error("http.internal_error", "path", r.URL.Path, "error", message,
			"request_id", trace.CausationID(r.Context()))
		message = "internal error"
	}
	writeJSON(w, status, map[string]any{"error": map[string]string{"code": code, "message": message}})
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}
