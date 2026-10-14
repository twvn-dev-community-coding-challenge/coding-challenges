package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/dotdak/sms-otp/internal/carrier"
	"github.com/dotdak/sms-otp/internal/providers"
	"github.com/dotdak/sms-otp/internal/ratelimit"
	"github.com/dotdak/sms-otp/internal/sms"
	"github.com/dotdak/sms-otp/models"
	"github.com/dotdak/sms-otp/repository"
	"github.com/labstack/echo/v4"
	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"
)

func TestRegister_BindError(t *testing.T) {
	e := echo.New()
	h := NewAuthHandler(nil, &stubSMS{}, nil)
	req := httptest.NewRequest(http.MethodPost, "/register", strings.NewReader("{"))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	if err := h.Register(c); err != nil {
		t.Fatal(err)
	}
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("code=%d", rec.Code)
	}
}

func TestRegister_Validation(t *testing.T) {
	e := echo.New()
	h := NewAuthHandler(nil, &stubSMS{}, nil)
	body := `{"username":"","password":"","phone_number":"","country":""}`
	req := httptest.NewRequest(http.MethodPost, "/register", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	if err := h.Register(c); err != nil {
		t.Fatal(err)
	}
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("code=%d", rec.Code)
	}
}

func TestRegister_InvalidCountry(t *testing.T) {
	e := echo.New()
	h := NewAuthHandler(nil, &stubSMS{}, nil)
	body := `{"username":"u1","password":"p1","phone_number":"6391712345678","country":"XX"}`
	req := httptest.NewRequest(http.MethodPost, "/register", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	if err := h.Register(c); err != nil {
		t.Fatal(err)
	}
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("code=%d", rec.Code)
	}
}

func TestRegister_HappyPath(t *testing.T) {
	db := testGormDB(t)
	userRepo := repository.NewUserRepository(db)
	smsRepo := repository.NewSMSGormRepository(db)
	resolver := carrier.NewPrefixCarrierResolver()
	router := providers.NewSimpleProviderRouter()
	router.RegisterRouter("PH", providers.CarrierGlobe, providers.ProviderMessageBird, &providers.MockProvider{Name: providers.ProviderMessageBird, SkipDelay: true, Quiet: true})
	smsSvc := sms.NewSMSServiceWithImmediateDispatch(smsRepo, resolver, router, nil)

	e := echo.New()
	h := NewAuthHandler(userRepo, smsSvc, nil)
	body := `{"username":"reguser","password":"secret12","phone_number":"6391712345999","country":"PH"}`
	req := httptest.NewRequest(http.MethodPost, "/register", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	if err := h.Register(c); err != nil {
		t.Fatal(err)
	}
	if rec.Code != http.StatusCreated {
		t.Fatalf("code=%d body=%s", rec.Code, rec.Body.String())
	}
}

func TestVerifyOTP_NotFound(t *testing.T) {
	db := testGormDB(t)
	userRepo := repository.NewUserRepository(db)
	e := echo.New()
	h := NewAuthHandler(userRepo, &stubSMS{}, nil)
	body := `{"username":"nope","otp_code":"123456"}`
	req := httptest.NewRequest(http.MethodPost, "/verify", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	if err := h.VerifyOTP(c); err != nil {
		t.Fatal(err)
	}
	if rec.Code != http.StatusNotFound {
		t.Fatalf("code=%d", rec.Code)
	}
}

func TestVerifyOTP_HappyPath(t *testing.T) {
	db := testGormDB(t)
	ctx := context.Background()
	userRepo := repository.NewUserRepository(db)
	hash, _ := bcrypt.GenerateFromPassword([]byte("pw"), bcrypt.MinCost)
	exp := time.Now().Add(time.Hour)
	u := &models.User{
		Username: "vuser", PasswordHash: string(hash), PhoneNumber: "6391711111111",
		Country: "PH", Status: models.UserStatusPending, OTPCode: "111111", OTPExpiresAt: &exp,
	}
	if err := userRepo.Create(ctx, u); err != nil {
		t.Fatal(err)
	}

	e := echo.New()
	h := NewAuthHandler(userRepo, &stubSMS{}, nil)
	payload := map[string]string{"username": "vuser", "otp_code": "111111"}
	raw, _ := json.Marshal(payload)
	req := httptest.NewRequest(http.MethodPost, "/verify", bytes.NewReader(raw))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	if err := h.VerifyOTP(c); err != nil {
		t.Fatal(err)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("code=%d body=%s", rec.Code, rec.Body.String())
	}
}

func TestLogin_BindError(t *testing.T) {
	e := echo.New()
	h := NewAuthHandler(nil, &stubSMS{}, nil)
	req := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader("{"))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	if err := h.Login(c); err != nil {
		t.Fatal(err)
	}
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("code=%d", rec.Code)
	}
}

func TestLogin_Unauthorized(t *testing.T) {
	db := testGormDB(t)
	userRepo := repository.NewUserRepository(db)
	e := echo.New()
	h := NewAuthHandler(userRepo, &stubSMS{}, nil)
	body := `{"username":"x","password":"y"}`
	req := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	if err := h.Login(c); err != nil {
		t.Fatal(err)
	}
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("code=%d", rec.Code)
	}
}

func TestLoginVerify_BadBody(t *testing.T) {
	e := echo.New()
	h := NewAuthHandler(nil, &stubSMS{}, nil)
	req := httptest.NewRequest(http.MethodPost, "/lv", strings.NewReader("{"))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	if err := h.LoginVerify(c); err != nil {
		t.Fatal(err)
	}
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("code=%d", rec.Code)
	}
}

func TestGetOTPTTL_env(t *testing.T) {
	t.Setenv("OTP_TTL_SECONDS", "120")
	// getOTPTTL reads env at call time
	if d := getOTPTTL(); d != 120*time.Second {
		t.Fatalf("got %v", d)
	}
	t.Setenv("OTP_TTL_SECONDS", "nope")
	if d := getOTPTTL(); d != defaultOTPTTL {
		t.Fatalf("invalid env should default, got %v", d)
	}
}

func TestRegister_RateLimit429(t *testing.T) {
	s, err := miniredis.Run()
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	rdb := redis.NewClient(&redis.Options{Addr: s.Addr()})
	defer rdb.Close()
	lim := ratelimit.New(rdb)

	db := testGormDB(t)
	userRepo := repository.NewUserRepository(db)
	h := NewAuthHandler(userRepo, &stubSMS{}, lim)

	e := echo.New()
	ip := "10.0.0.50"
	for i := 0; i < regStartsPerIPPerHour; i++ {
		body := fmt.Sprintf(`{"username":"rlu%d","password":"pw123456","phone_number":"6391712345%03d","country":"PH"}`, i, i)
		req := httptest.NewRequest(http.MethodPost, "/register", strings.NewReader(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		req.Header.Set("X-Forwarded-For", ip)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		if err := h.Register(c); err != nil {
			t.Fatalf("iter %d: %v", i, err)
		}
		if rec.Code != http.StatusCreated {
			t.Fatalf("iter %d: code=%d body=%s", i, rec.Code, rec.Body.String())
		}
	}
	body := `{"username":"rlufinal","password":"pw123456","phone_number":"6391712345998","country":"PH"}`
	req := httptest.NewRequest(http.MethodPost, "/register", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req.Header.Set("X-Forwarded-For", ip)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	if err := h.Register(c); err != nil {
		t.Fatal(err)
	}
	if rec.Code != http.StatusTooManyRequests {
		t.Fatalf("want 429 got %d body=%s", rec.Code, rec.Body.String())
	}
}
