package app_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"ridedispatch/internal/app"
	"ridedispatch/internal/platform/clock"
)

var t0 = time.Date(2026, 10, 31, 9, 0, 0, 0, time.UTC)

const rideBody = `{
  "passengerId": "P001",
  "pickup": {"latitude": 10.7769, "longitude": 106.7009},
  "destination": {"latitude": 10.8231, "longitude": 106.6297}
}`

// client drives the fully wired application (seed data included) through
// its HTTP handler.
type client struct {
	t     *testing.T
	app   *app.App
	clock *clock.Fake
	keys  int
}

func newClient(t *testing.T) *client {
	t.Helper()
	clk := clock.NewFake(t0)
	a, err := app.New(app.Config{RadiusKm: 2, OfferTimeout: 30 * time.Second, Workers: 4}, clk, nil)
	if err != nil {
		t.Fatal(err)
	}
	a.Start()
	t.Cleanup(a.Stop)
	return &client{t: t, app: a, clock: clk}
}

type response struct {
	*httptest.ResponseRecorder
	t *testing.T
}

func (c *client) do(method, path, body string, headers ...string) response {
	c.t.Helper()
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	for i := 0; i+1 < len(headers); i += 2 {
		req.Header.Set(headers[i], headers[i+1])
	}
	rec := httptest.NewRecorder()
	c.app.Handler.ServeHTTP(rec, req)
	return response{rec, c.t}
}

func (c *client) settle() {
	c.t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := c.app.Bus.WaitIdle(ctx); err != nil {
		c.t.Fatal(err)
	}
}

// createRide posts the brief's example request and waits for dispatch.
func (c *client) createRide() string {
	c.t.Helper()
	c.keys++
	res := c.do("POST", "/rides", rideBody, "Idempotency-Key", "key-"+string(rune('a'+c.keys)))
	res.expect(http.StatusCreated)
	c.settle()
	return res.json()["rideId"].(string)
}

func (r response) expect(status int) response {
	r.t.Helper()
	if r.Code != status {
		r.t.Fatalf("status = %d, want %d; body: %s", r.Code, status, r.Body.String())
	}
	return r
}

func (r response) json() map[string]any {
	r.t.Helper()
	var m map[string]any
	if err := json.Unmarshal(r.Body.Bytes(), &m); err != nil {
		r.t.Fatalf("invalid JSON %q: %v", r.Body.String(), err)
	}
	return m
}

func (r response) errorCode() string {
	r.t.Helper()
	e, _ := r.json()["error"].(map[string]any)
	code, _ := e["code"].(string)
	return code
}

func (r response) field(name string) any {
	r.t.Helper()
	return r.json()[name]
}

func TestCreateAndGetRide(t *testing.T) {
	c := newClient(t)
	res := c.do("POST", "/rides", rideBody, "Idempotency-Key", "K1").expect(http.StatusCreated)
	body := res.json()
	id, _ := body["rideId"].(string)
	if id == "" || body["status"] != "REQUESTED" || body["passengerId"] != "P001" || body["assignedDriverId"] != nil {
		t.Fatalf("created ride = %v", body)
	}
	pickup := body["pickup"].(map[string]any)
	if pickup["latitude"] != 10.7769 || pickup["longitude"] != 106.7009 {
		t.Fatalf("pickup = %v", pickup)
	}
	if res.Header().Get("Location") != "/rides/"+id {
		t.Fatalf("Location header = %q", res.Header().Get("Location"))
	}

	c.settle()
	got := c.do("GET", "/rides/"+id, "").expect(http.StatusOK).json()
	if got["status"] != "DRIVER_ASSIGNED" || got["assignedDriverId"] != "D001" {
		t.Fatalf("after dispatch ride = %v", got)
	}
}

func TestCreateRideValidation(t *testing.T) {
	c := newClient(t)
	cases := []struct {
		name   string
		body   string
		key    string
		status int
		code   string
	}{
		{"missing key", rideBody, "", 400, "IDEMPOTENCY_KEY_MISSING"},
		{"malformed JSON", `{"passengerId":`, "K1", 400, "VALIDATION_FAILED"},
		{"missing passenger", strings.Replace(rideBody, `"P001"`, `""`, 1), "K2", 400, "VALIDATION_FAILED"},
		{"unknown passenger", strings.Replace(rideBody, `"P001"`, `"P999"`, 1), "K3", 404, "NOT_FOUND"},
		{"latitude out of range", strings.Replace(rideBody, "10.7769", "91", 1), "K4", 400, "VALIDATION_FAILED"},
		{"longitude out of range", strings.Replace(rideBody, "106.6297", "-180.5", 1), "K5", 400, "VALIDATION_FAILED"},
		{"missing pickup", `{"passengerId":"P001","destination":{"latitude":1,"longitude":1}}`, "K6", 400, "VALIDATION_FAILED"},
		{"missing longitude", `{"passengerId":"P001","pickup":{"latitude":1},"destination":{"latitude":1,"longitude":1}}`, "K7", 400, "VALIDATION_FAILED"},
		{"coordinate as string", strings.Replace(rideBody, "10.7769", `"10.7769"`, 1), "K8", 400, "VALIDATION_FAILED"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var res response
			if tc.key == "" {
				res = c.do("POST", "/rides", tc.body)
			} else {
				res = c.do("POST", "/rides", tc.body, "Idempotency-Key", tc.key)
			}
			res.expect(tc.status)
			if res.errorCode() != tc.code {
				t.Fatalf("error code = %q, want %q", res.errorCode(), tc.code)
			}
		})
	}
}

func TestCreateRideIdempotencyKey(t *testing.T) {
	c := newClient(t)
	first := c.do("POST", "/rides", rideBody, "Idempotency-Key", "K1").expect(http.StatusCreated)
	if first.Header().Get("Idempotent-Replayed") != "" {
		t.Fatal("first request marked as replay")
	}
	c.settle()

	// Same key, same request (whitespace differs): the original result, no new ride.
	compact := strings.Join(strings.Fields(rideBody), "")
	replay := c.do("POST", "/rides", compact, "Idempotency-Key", "K1").expect(http.StatusCreated)
	if replay.Header().Get("Idempotent-Replayed") != "true" {
		t.Fatal("replay not marked with Idempotent-Replayed: true")
	}
	if replay.field("rideId") != first.field("rideId") || replay.field("status") != "REQUESTED" {
		t.Fatalf("replay = %s, want original %s", replay.Body.String(), first.Body.String())
	}

	// Same key, different request.
	other := strings.Replace(rideBody, "10.8231", "10.9", 1)
	res := c.do("POST", "/rides", other, "Idempotency-Key", "K1").expect(http.StatusUnprocessableEntity)
	if res.errorCode() != "IDEMPOTENCY_KEY_REUSED" {
		t.Fatalf("error code = %q", res.errorCode())
	}

	// Only one ride was dispatched: D002 is still free.
	if c.do("GET", "/drivers/D002", "").expect(http.StatusOK).field("status") != "AVAILABLE" {
		t.Fatal("replay created a second ride")
	}
}

func TestGetUnknownRide(t *testing.T) {
	c := newClient(t)
	if code := c.do("GET", "/rides/R-missing", "").expect(http.StatusNotFound).errorCode(); code != "NOT_FOUND" {
		t.Fatalf("error code = %q", code)
	}
	c.do("GET", "/rides/R-missing/timeline", "").expect(http.StatusNotFound)
}

func TestDriverAvailability(t *testing.T) {
	c := newClient(t)
	res := c.do("PUT", "/drivers/D004/availability", `{"status":"AVAILABLE"}`).expect(http.StatusOK)
	if res.field("status") != "AVAILABLE" || res.field("driverId") != "D004" {
		t.Fatalf("driver = %s", res.Body.String())
	}
	c.do("PUT", "/drivers/D004/availability", `{"status":"OFFLINE"}`).expect(http.StatusOK)
	c.do("PUT", "/drivers/D004/availability", `{"status":"BUSY"}`).expect(http.StatusBadRequest)
	c.do("PUT", "/drivers/D004/availability", `{}`).expect(http.StatusBadRequest)
	c.do("PUT", "/drivers/D999/availability", `{"status":"AVAILABLE"}`).expect(http.StatusNotFound)

	// An OFFERED driver cannot go offline.
	c.createRide()
	res = c.do("PUT", "/drivers/D001/availability", `{"status":"OFFLINE"}`).expect(http.StatusConflict)
	if res.errorCode() != "INVALID_TRANSITION" {
		t.Fatalf("error code = %q", res.errorCode())
	}
}

func TestDriverLocation(t *testing.T) {
	c := newClient(t)
	c.clock.Advance(time.Minute)
	res := c.do("PUT", "/drivers/D003/location", `{"latitude":10.7770,"longitude":106.7010}`).expect(http.StatusOK)
	loc := res.field("location").(map[string]any)
	if loc["latitude"] != 10.777 || loc["longitude"] != 106.701 {
		t.Fatalf("location = %v", loc)
	}
	if res.field("locationUpdatedAt") != t0.Add(time.Minute).Format(time.RFC3339) {
		t.Fatalf("locationUpdatedAt = %v", res.field("locationUpdatedAt"))
	}
	c.do("PUT", "/drivers/D003/location", `{"latitude":95,"longitude":106.7}`).expect(http.StatusBadRequest)
	c.do("PUT", "/drivers/D003/location", `{"latitude":10.7}`).expect(http.StatusBadRequest)
	c.do("PUT", "/drivers/D999/location", `{"latitude":10.7,"longitude":106.7}`).expect(http.StatusNotFound)

	// Matching uses the latest location: D003 moved next to the pickup.
	id := c.createRide()
	if got := c.do("GET", "/rides/"+id, "").json()["assignedDriverId"]; got != "D003" {
		t.Fatalf("assigned to %v, want D003 at its new location", got)
	}
}

func TestOfferEndpoints(t *testing.T) {
	c := newClient(t)
	id := c.createRide()

	c.do("POST", "/rides/"+id+"/offers/D002/accept", "").expect(http.StatusNotFound)
	c.do("POST", "/rides/"+id+"/complete", "").expect(http.StatusConflict)

	res := c.do("POST", "/rides/"+id+"/offers/D001/reject", "").expect(http.StatusOK)
	if res.field("status") != "MATCHING" || res.field("assignedDriverId") != nil {
		t.Fatalf("after reject = %s", res.Body.String())
	}
	c.settle()
	if got := c.do("GET", "/rides/"+id, "").json()["assignedDriverId"]; got != "D002" {
		t.Fatalf("rematched to %v, want D002", got)
	}
	c.do("POST", "/rides/"+id+"/offers/D001/accept", "").expect(http.StatusConflict)

	res = c.do("POST", "/rides/"+id+"/offers/D002/accept", "").expect(http.StatusOK)
	if res.field("status") != "DRIVER_ACCEPTED" {
		t.Fatalf("after accept = %s", res.Body.String())
	}
	c.do("POST", "/rides/"+id+"/offers/D002/accept", "").expect(http.StatusOK)
	if c.do("GET", "/drivers/D002", "").field("status") != "BUSY" {
		t.Fatal("accepted driver not BUSY")
	}

	res = c.do("POST", "/rides/"+id+"/complete", "").expect(http.StatusOK)
	if res.field("status") != "COMPLETED" {
		t.Fatalf("after complete = %s", res.Body.String())
	}
	if c.do("GET", "/drivers/D002", "").field("status") != "AVAILABLE" {
		t.Fatal("driver not released after completion")
	}
	if code := c.do("POST", "/rides/"+id+"/cancel", "").expect(http.StatusConflict).errorCode(); code != "INVALID_TRANSITION" {
		t.Fatalf("cancel completed ride code = %q", code)
	}
}

func TestCancelEndpoint(t *testing.T) {
	c := newClient(t)
	id := c.createRide()
	res := c.do("POST", "/rides/"+id+"/cancel", "").expect(http.StatusOK)
	if res.field("status") != "CANCELLED" {
		t.Fatalf("after cancel = %s", res.Body.String())
	}
	c.do("POST", "/rides/"+id+"/cancel", "").expect(http.StatusOK)
	if c.do("GET", "/drivers/D001", "").field("status") != "AVAILABLE" {
		t.Fatal("driver not released after cancel")
	}
	c.do("POST", "/rides/R-missing/cancel", "").expect(http.StatusNotFound)
}

func TestOfferTimeoutThroughTheAPI(t *testing.T) {
	c := newClient(t)
	id := c.createRide()
	c.clock.Advance(30 * time.Second)
	c.settle()
	if got := c.do("GET", "/rides/"+id, "").json()["assignedDriverId"]; got != "D002" {
		t.Fatalf("after timeout assigned to %v, want D002", got)
	}
	if c.do("GET", "/drivers/D001", "").field("status") != "AVAILABLE" {
		t.Fatal("timed-out driver not released")
	}
}

func TestTimelineTellsWhatHappenedToARide(t *testing.T) {
	c := newClient(t)
	res := c.do("POST", "/rides", rideBody, "Idempotency-Key", "K1", "X-Request-ID", "req-123").expect(http.StatusCreated)
	if res.Header().Get("X-Request-ID") != "req-123" {
		t.Fatal("request id not echoed")
	}
	id := res.field("rideId").(string)
	c.settle()
	c.do("POST", "/rides/"+id+"/offers/D001/reject", "").expect(http.StatusOK)
	c.settle()
	c.do("POST", "/rides/"+id+"/offers/D002/accept", "").expect(http.StatusOK)
	c.settle()

	body := c.do("GET", "/rides/"+id+"/timeline", "").expect(http.StatusOK).json()
	if body["rideId"] != id {
		t.Fatalf("timeline rideId = %v", body["rideId"])
	}
	var types []string
	evs := body["events"].([]any)
	for _, raw := range evs {
		types = append(types, raw.(map[string]any)["type"].(string)+":"+stringOr(raw.(map[string]any)["driverId"]))
	}
	want := []string{"RideCreated:", "OfferCreated:D001", "OfferRejected:D001", "OfferCreated:D002", "OfferAccepted:D002"}
	if strings.Join(types, ",") != strings.Join(want, ",") {
		t.Fatalf("timeline = %v, want %v", types, want)
	}
	first := evs[0].(map[string]any)
	if first["causationId"] != "req-123" {
		t.Fatalf("RideCreated causationId = %v, want the request id", first["causationId"])
	}
	if evs[1].(map[string]any)["causationId"] != first["eventId"] {
		t.Fatal("OfferCreated is not linked to RideCreated")
	}
}

func stringOr(v any) string {
	s, _ := v.(string)
	return s
}
