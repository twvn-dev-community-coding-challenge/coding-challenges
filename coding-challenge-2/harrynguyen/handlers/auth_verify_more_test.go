package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/dotdak/sms-otp/models"
	"github.com/dotdak/sms-otp/repository"
	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"
)

func TestVerifyOTP_Validation(t *testing.T) {
	e := echo.New()
	h := NewAuthHandler(repository.NewUserRepository(testGormDB(t)), &stubSMS{}, nil)
	req := httptest.NewRequest(http.MethodPost, "/verify", bytes.NewReader([]byte(`{"username":"","otp_code":""}`)))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	if err := h.VerifyOTP(c); err != nil {
		t.Fatal(err)
	}
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("code=%d", rec.Code)
	}
}

func TestVerifyOTP_AlreadyActive(t *testing.T) {
	db := testGormDB(t)
	ctx := context.Background()
	repo := repository.NewUserRepository(db)
	hash, _ := bcrypt.GenerateFromPassword([]byte("pw"), bcrypt.MinCost)
	if err := repo.Create(ctx, &models.User{
		Username: "act", PasswordHash: string(hash), PhoneNumber: "6391712222001",
		Country: "PH", Status: models.UserStatusActive, OTPCode: "111111",
	}); err != nil {
		t.Fatal(err)
	}
	e := echo.New()
	h := NewAuthHandler(repo, &stubSMS{}, nil)
	raw, _ := json.Marshal(map[string]string{"username": "act", "otp_code": "111111"})
	req := httptest.NewRequest(http.MethodPost, "/verify", bytes.NewReader(raw))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	if err := h.VerifyOTP(c); err != nil {
		t.Fatal(err)
	}
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("code=%d", rec.Code)
	}
}

func TestVerifyOTP_WrongOTP(t *testing.T) {
	db := testGormDB(t)
	ctx := context.Background()
	repo := repository.NewUserRepository(db)
	hash, _ := bcrypt.GenerateFromPassword([]byte("pw"), bcrypt.MinCost)
	exp := time.Now().Add(time.Hour)
	if err := repo.Create(ctx, &models.User{
		Username: "wotp", PasswordHash: string(hash), PhoneNumber: "6391712222002",
		Country: "PH", Status: models.UserStatusPending, OTPCode: "222222", OTPExpiresAt: &exp,
	}); err != nil {
		t.Fatal(err)
	}
	e := echo.New()
	h := NewAuthHandler(repo, &stubSMS{}, nil)
	raw, _ := json.Marshal(map[string]string{"username": "wotp", "otp_code": "000000"})
	req := httptest.NewRequest(http.MethodPost, "/verify", bytes.NewReader(raw))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	if err := h.VerifyOTP(c); err != nil {
		t.Fatal(err)
	}
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("code=%d", rec.Code)
	}
}

func TestVerifyOTP_Expired(t *testing.T) {
	db := testGormDB(t)
	ctx := context.Background()
	repo := repository.NewUserRepository(db)
	hash, _ := bcrypt.GenerateFromPassword([]byte("pw"), bcrypt.MinCost)
	exp := time.Now().Add(-time.Hour)
	if err := repo.Create(ctx, &models.User{
		Username: "exp", PasswordHash: string(hash), PhoneNumber: "6391712222003",
		Country: "PH", Status: models.UserStatusPending, OTPCode: "333333", OTPExpiresAt: &exp,
	}); err != nil {
		t.Fatal(err)
	}
	e := echo.New()
	h := NewAuthHandler(repo, &stubSMS{}, nil)
	raw, _ := json.Marshal(map[string]string{"username": "exp", "otp_code": "333333"})
	req := httptest.NewRequest(http.MethodPost, "/verify", bytes.NewReader(raw))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	if err := h.VerifyOTP(c); err != nil {
		t.Fatal(err)
	}
	if rec.Code != http.StatusGone {
		t.Fatalf("code=%d", rec.Code)
	}
}
