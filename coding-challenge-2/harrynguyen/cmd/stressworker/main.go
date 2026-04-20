// Stress worker: endless mixed load against the SMS OTP API (registration, verify, internal SMS,
// provider callbacks, random x-test-sms-modes). Requires server ALLOW_TEST_SMS_HEADERS=true.
//
// Run: SMS_INTERNAL_API_KEY=dev-local-sms-key go run ./cmd/stressworker -base http://127.0.0.1:8080
//
// Optional: STRESS_WORKERS (default 4) sets concurrent scenario goroutines; -workers still overrides the default from env.
//
// Optional: SMS_CALLBACK_SECRET if the API enforces webhook secret.
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"time"
)

var (
	statIterations    atomic.Uint64
	statRegOK         atomic.Uint64
	statRegFail       atomic.Uint64
	statVerifyOK      atomic.Uint64
	statVerifyFail    atomic.Uint64
	statSendSMSOK     atomic.Uint64
	statSendSMSFail   atomic.Uint64
	statCallbackOK    atomic.Uint64
	statCallbackFail  atomic.Uint64
	statRecoveryNoise atomic.Uint64
	statHTTP429       atomic.Uint64
)

var otpRe = regexp.MustCompile(`(\d{6})`)

type config struct {
	base           string
	internalKey    string
	callbackSecret string
	client         *http.Client
	maxPause       time.Duration
}

func main() {
	base := flag.String("base", envOr("STRESS_BASE_URL", "http://127.0.0.1:8080"), "API base URL")
	workers := flag.Int("workers", envIntOr("STRESS_WORKERS", 4), "concurrent scenario goroutines (env STRESS_WORKERS)")
	maxPause := flag.Duration("max-pause", 400*time.Millisecond, "max random pause between scenario steps")
	statsEvery := flag.Duration("stats-every", 30*time.Second, "log rolling stats interval")
	flag.Parse()

	internalKey := strings.TrimSpace(os.Getenv("SMS_INTERNAL_API_KEY"))
	if internalKey == "" {
		internalKey = strings.TrimSpace(os.Getenv("PW_SMS_INTERNAL_API_KEY"))
	}
	callbackSecret := strings.TrimSpace(os.Getenv("SMS_CALLBACK_SECRET"))

	if internalKey == "" {
		log.Println("warn: SMS_INTERNAL_API_KEY unset — internal /api/sms/send and /messages scenarios will be skipped")
	}

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	client := &http.Client{Timeout: 45 * time.Second}
	cfg := config{
		base:           strings.TrimRight(*base, "/"),
		internalKey:    internalKey,
		callbackSecret: callbackSecret,
		client:         client,
		maxPause:       *maxPause,
	}

	go statsLogger(ctx, *statsEvery)

	var wg sync.WaitGroup
	for i := 0; i < *workers; i++ {
		wg.Add(1)
		go func(seed int64) {
			defer wg.Done()
			rng := rand.New(rand.NewSource(time.Now().UnixNano() ^ seed))
			workerLoop(ctx, cfg, rng)
		}(int64(i))
	}
	<-ctx.Done()
	log.Println("shutting down, waiting workers...")
	wg.Wait()
	log.Println("bye")
}

func workerLoop(ctx context.Context, cfg config, rng *rand.Rand) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}
		statIterations.Add(1)
		n := rng.Intn(100)
		switch {
		case n < 12 && cfg.internalKey != "":
			// Drives sms_otp_sms_recovery_events_total for Grafana (Queue→delivered recovery + mitigation).
			scenarioRecoveryMetricsProbe(ctx, cfg, rng)
		case n < 40:
			scenarioRegisterVerify(ctx, cfg, rng, false, "")
		case n < 52:
			scenarioRegisterVerify(ctx, cfg, rng, true, "transient-failure-once")
		case n < 60:
			scenarioRegisterVerify(ctx, cfg, rng, true, "always-success")
		case n < 67:
			scenarioRegisterAlwaysFail(ctx, cfg, rng)
		case n < 70:
			scenarioWrongOTP(ctx, cfg, rng)
		case n < 87 && cfg.internalKey != "":
			scenarioInternalSendAndCallbacks(ctx, cfg, rng)
		case n < 94 && cfg.internalKey != "":
			scenarioCallbackStorm(ctx, cfg, rng)
		case cfg.internalKey != "":
			scenarioRandomCallbackGarbage(ctx, cfg, rng)
		default:
			scenarioRegisterVerify(ctx, cfg, rng, false, "")
		}
		jitterSleep(ctx, cfg, rng)
	}
}

func jitterSleep(ctx context.Context, cfg config, rng *rand.Rand) {
	if cfg.maxPause <= 0 {
		return
	}
	d := time.Duration(rng.Int63n(int64(cfg.maxPause)))
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-ctx.Done():
	case <-t.C:
	}
}

func scenarioRegisterVerify(ctx context.Context, cfg config, rng *rand.Rand, testHeader bool, mode string) {
	username := fmt.Sprintf("stress_%d_%d", time.Now().UnixNano(), rng.Intn(1_000_000))
	phone := fmt.Sprintf("63917%07d", rng.Intn(10_000_000))
	body := map[string]any{
		"username":     username,
		"password":     "Str3ss!Pass9",
		"phone_number": phone,
		"country":      "PH",
	}
	h := headers(cfg, rng)
	if testHeader && mode != "" {
		h["X-Test-Sms-Mode"] = mode
	}
	status, resp, err := doJSON(ctx, cfg, http.MethodPost, "/api/register", body, h)
	if err != nil {
		statRegFail.Add(1)
		return
	}
	if status == http.StatusTooManyRequests {
		statHTTP429.Add(1)
		statRegFail.Add(1)
		return
	}
	if status != http.StatusCreated {
		statRegFail.Add(1)
		statRecoveryNoise.Add(1)
		return
	}
	statRegOK.Add(1)

	var envelope struct {
		User struct {
			Username string `json:"username"`
		} `json:"user"`
	}
	if json.Unmarshal(resp, &envelope) != nil || envelope.User.Username == "" {
		return
	}
	uname := envelope.User.Username

	jitterSleep(ctx, cfg, rng)

	if cfg.internalKey == "" {
		return
	}
	otp, ok := fetchOTPForPhone(ctx, cfg, phone)
	if !ok {
		statVerifyFail.Add(1)
		return
	}
	vbody := map[string]any{"username": uname, "otp_code": otp}
	vstatus, _, err := doJSON(ctx, cfg, http.MethodPost, "/api/verify", vbody, headers(cfg, rng))
	if err != nil {
		statVerifyFail.Add(1)
		return
	}
	if vstatus == http.StatusOK {
		statVerifyOK.Add(1)
	} else {
		statVerifyFail.Add(1)
	}
}

func scenarioRegisterAlwaysFail(ctx context.Context, cfg config, rng *rand.Rand) {
	username := fmt.Sprintf("fail_%d_%d", time.Now().UnixNano(), rng.Intn(1_000_000))
	phone := fmt.Sprintf("63917%07d", rng.Intn(10_000_000))
	body := map[string]any{
		"username": username, "password": "Str3ss!Pass9",
		"phone_number": phone, "country": "PH",
	}
	h := headers(cfg, rng)
	h["X-Test-Sms-Mode"] = "always-fail"
	status, _, err := doJSON(ctx, cfg, http.MethodPost, "/api/register", body, h)
	if err != nil {
		statRegFail.Add(1)
		return
	}
	if status == http.StatusInternalServerError {
		statRecoveryNoise.Add(1)
		statRegFail.Add(1)
		return
	}
	if status == http.StatusTooManyRequests {
		statHTTP429.Add(1)
	}
	statRegFail.Add(1)
}

func scenarioWrongOTP(ctx context.Context, cfg config, rng *rand.Rand) {
	username := fmt.Sprintf("badotp_%d_%d", time.Now().UnixNano(), rng.Intn(1_000_000))
	phone := fmt.Sprintf("63917%07d", rng.Intn(10_000_000))
	body := map[string]any{
		"username": username, "password": "Str3ss!Pass9",
		"phone_number": phone, "country": "PH",
	}
	status, _, err := doJSON(ctx, cfg, http.MethodPost, "/api/register", body, headers(cfg, rng))
	if err != nil || status != http.StatusCreated {
		return
	}
	vbody := map[string]any{"username": username, "otp_code": "000000"}
	vstatus, _, _ := doJSON(ctx, cfg, http.MethodPost, "/api/verify", vbody, headers(cfg, rng))
	if vstatus == http.StatusUnauthorized {
		statRecoveryNoise.Add(1)
		statVerifyFail.Add(1)
	}
}

// scenarioRecoveryMetricsProbe issues a successful internal send (Queue) then callbacks that hit
// applyStatusWithRecovery: Send-success from Queue, then Send-failed after terminal success (mitigation).
func scenarioRecoveryMetricsProbe(ctx context.Context, cfg config, rng *rand.Rand) {
	phone := fmt.Sprintf("63917%07d", rng.Intn(10_000_000))
	body := map[string]any{
		"country": "PH", "phone_number": phone,
		"content": fmt.Sprintf("recovery-probe %d", time.Now().UnixNano()),
	}
	h := headers(cfg, rng)
	h["Authorization"] = "Bearer " + cfg.internalKey
	status, resp, err := doJSON(ctx, cfg, http.MethodPost, "/api/sms/send", body, h)
	if err != nil || status != http.StatusAccepted {
		return
	}
	var sendOut struct {
		ID string `json:"id"`
	}
	if json.Unmarshal(resp, &sendOut) != nil || sendOut.ID == "" {
		return
	}
	pid, ok := waitForSMSProviderMessageID(ctx, cfg, rng, sendOut.ID, 15*time.Second)
	if !ok || pid == "" {
		return
	}
	fireCallback(ctx, cfg, pid, "Send-success", 0.01)
	jitterSleep(ctx, cfg, rng)
	fireCallback(ctx, cfg, pid, "Send-failed", 0.02)
}

func scenarioInternalSendAndCallbacks(ctx context.Context, cfg config, rng *rand.Rand) {
	modes := []string{"", "transient-failure-once", "always-success", "always-fail"}
	mode := modes[rng.Intn(len(modes))]
	phone := fmt.Sprintf("63917%07d", rng.Intn(10_000_000))
	body := map[string]any{
		"country": "PH", "phone_number": phone,
		"content": fmt.Sprintf("stress internal %d", time.Now().UnixNano()),
	}
	h := headers(cfg, rng)
	h["Authorization"] = "Bearer " + cfg.internalKey
	if mode != "" {
		h["X-Test-Sms-Mode"] = mode
	}
	status, resp, err := doJSON(ctx, cfg, http.MethodPost, "/api/sms/send", body, h)
	if err != nil {
		statSendSMSFail.Add(1)
		return
	}
	if status == http.StatusTooManyRequests {
		statHTTP429.Add(1)
		statSendSMSFail.Add(1)
		return
	}
	if status != http.StatusAccepted {
		statSendSMSFail.Add(1)
		statRecoveryNoise.Add(1)
		return
	}
	statSendSMSOK.Add(1)

	var sendOut struct {
		ID string `json:"id"`
	}
	if json.Unmarshal(resp, &sendOut) != nil || sendOut.ID == "" {
		return
	}
	providerID, ok := waitForSMSProviderMessageID(ctx, cfg, rng, sendOut.ID, 15*time.Second)
	if !ok || providerID == "" {
		return
	}

	// Chaotic callback sequences (recovery paths in HandleCallback).
	statuses := []string{
		"Send-success",
		"Send-failed",
		"Send-to-carrier",
		"Carrier-rejected",
	}
	for i := 0; i < 1+rng.Intn(6); i++ {
		jitterSleep(ctx, cfg, rng)
		st := statuses[rng.Intn(len(statuses))]
		fireCallback(ctx, cfg, providerID, st, rng.Float64()*0.02)
	}
}

func scenarioCallbackStorm(ctx context.Context, cfg config, rng *rand.Rand) {
	phone := fmt.Sprintf("63917%07d", rng.Intn(10_000_000))
	body := map[string]any{
		"country": "PH", "phone_number": phone,
		"content": fmt.Sprintf("storm %d", time.Now().UnixNano()),
	}
	h := headers(cfg, rng)
	h["Authorization"] = "Bearer " + cfg.internalKey
	status, resp, err := doJSON(ctx, cfg, http.MethodPost, "/api/sms/send", body, h)
	if err != nil || status != http.StatusAccepted {
		return
	}
	var sendOut struct {
		ID string `json:"id"`
	}
	if json.Unmarshal(resp, &sendOut) != nil || sendOut.ID == "" {
		return
	}
	providerID, ok := waitForSMSProviderMessageID(ctx, cfg, rng, sendOut.ID, 15*time.Second)
	if !ok || providerID == "" {
		return
	}
	var wg sync.WaitGroup
	n := 8 + rng.Intn(25)
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			fireCallback(ctx, cfg, providerID, "Send-success", 0.01)
		}()
	}
	wg.Wait()
}

func scenarioRandomCallbackGarbage(ctx context.Context, cfg config, rng *rand.Rand) {
	// Random provider IDs — exercises error handling / logs without assuming state.
	id := fmt.Sprintf("fake-%d-%d", time.Now().UnixNano(), rng.Int())
	st := []string{"Send-success", "Send-failed", "Queue"}[rng.Intn(3)]
	fireCallback(ctx, cfg, id, st, 0)
	statRecoveryNoise.Add(1)
}

func fireCallback(ctx context.Context, cfg config, messageID, status string, cost float64) {
	body := map[string]any{
		"message_id":  messageID,
		"status":      status,
		"actual_cost": cost,
	}
	h := map[string]string{}
	if cfg.callbackSecret != "" {
		h["X-Callback-Secret"] = cfg.callbackSecret
	}
	statusCode, _, err := doJSON(ctx, cfg, http.MethodPost, "/api/sms/callback", body, h)
	if err != nil {
		statCallbackFail.Add(1)
		return
	}
	if statusCode == http.StatusOK {
		statCallbackOK.Add(1)
	} else {
		statCallbackFail.Add(1)
	}
}

func fetchOTPForPhone(ctx context.Context, cfg config, phone string) (string, bool) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, cfg.base+"/api/sms/messages", nil)
	if err != nil {
		return "", false
	}
	req.Header.Set("Authorization", "Bearer "+cfg.internalKey)
	res, err := cfg.client.Do(req)
	if err != nil {
		return "", false
	}
	defer res.Body.Close()
	b, err := io.ReadAll(res.Body)
	if err != nil || res.StatusCode != http.StatusOK {
		return "", false
	}
	var msgs []struct {
		PhoneNumber string `json:"phone_number"`
		Content     string `json:"content"`
	}
	if json.Unmarshal(b, &msgs) != nil {
		return "", false
	}
	for i := len(msgs) - 1; i >= 0; i-- {
		m := msgs[i]
		if m.PhoneNumber == phone {
			if sm := otpRe.FindStringSubmatch(m.Content); len(sm) > 1 {
				return sm[1], true
			}
		}
	}
	return "", false
}

func headers(cfg config, rng *rand.Rand) map[string]string {
	return map[string]string{
		"Content-Type":         "application/json",
		"X-Forwarded-For":      randomPrivateIP(rng),
		"X-Device-Fingerprint": fmt.Sprintf("stress-%d", rng.Intn(50_000)),
	}
}

func randomPrivateIP(rng *rand.Rand) string {
	return fmt.Sprintf("10.%d.%d.%d", rng.Intn(256), rng.Intn(256), rng.Intn(256))
}

// waitForSMSProviderMessageID polls GET /api/sms/messages/:id until the row has a provider message_id,
// the message reaches Send-failed (no provider id expected), or deadline.
func waitForSMSProviderMessageID(ctx context.Context, cfg config, rng *rand.Rand, smsID string, deadline time.Duration) (providerMsgID string, ok bool) {
	if smsID == "" {
		return "", false
	}
	deadlineAt := time.Now().Add(deadline)
	h := headers(cfg, rng)
	h["Authorization"] = "Bearer " + cfg.internalKey
	for time.Now().Before(deadlineAt) {
		select {
		case <-ctx.Done():
			return "", false
		default:
		}
		st, b, err := doJSON(ctx, cfg, http.MethodGet, "/api/sms/messages/"+smsID, nil, h)
		if err != nil || st != http.StatusOK {
			time.Sleep(50 * time.Millisecond)
			continue
		}
		var out struct {
			Message struct {
				MessageID string `json:"message_id"`
				Status    string `json:"status"`
			} `json:"message"`
		}
		if json.Unmarshal(b, &out) != nil {
			time.Sleep(50 * time.Millisecond)
			continue
		}
		if out.Message.MessageID != "" {
			return out.Message.MessageID, true
		}
		if out.Message.Status == "Send-failed" {
			return "", false
		}
		time.Sleep(50 * time.Millisecond)
	}
	return "", false
}

func doJSON(ctx context.Context, cfg config, method, path string, body any, extra map[string]string) (int, []byte, error) {
	var rdr io.Reader
	if body != nil {
		raw, err := json.Marshal(body)
		if err != nil {
			return 0, nil, err
		}
		rdr = bytes.NewReader(raw)
	}
	req, err := http.NewRequestWithContext(ctx, method, cfg.base+path, rdr)
	if err != nil {
		return 0, nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	for k, v := range extra {
		req.Header.Set(k, v)
	}
	res, err := cfg.client.Do(req)
	if err != nil {
		return 0, nil, err
	}
	defer res.Body.Close()
	b, err := io.ReadAll(res.Body)
	return res.StatusCode, b, err
}

func statsLogger(ctx context.Context, every time.Duration) {
	t := time.NewTicker(every)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			log.Printf(
				"stats iter=%d reg_ok=%d reg_fail=%d verify_ok=%d verify_fail=%d send_ok=%d send_fail=%d cb_ok=%d cb_fail=%d noise=%d http429=%d",
				statIterations.Load(),
				statRegOK.Load(), statRegFail.Load(),
				statVerifyOK.Load(), statVerifyFail.Load(),
				statSendSMSOK.Load(), statSendSMSFail.Load(),
				statCallbackOK.Load(), statCallbackFail.Load(),
				statRecoveryNoise.Load(),
				statHTTP429.Load(),
			)
		}
	}
}

func envOr(key, def string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return def
}

// envIntOr parses key as a base-10 int; on empty, invalid, or <1 returns def.
func envIntOr(key string, def int) int {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return def
	}
	n, err := strconv.Atoi(v)
	if err != nil || n < 1 {
		return def
	}
	return n
}
