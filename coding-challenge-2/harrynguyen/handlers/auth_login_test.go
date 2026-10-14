package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/dotdak/sms-otp/models"
	"github.com/dotdak/sms-otp/repository"
	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"
)

func TestLogin_Validation(t *testing.T) {
	e := echo.New()
	h := NewAuthHandler(nil, &stubSMS{}, nil)
	req := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(`{"username":"","password":""}`))
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

func TestLogin_HappyPath(t *testing.T) {
	db := testGormDB(t)
	ctx := context.Background()
	repo := repository.NewUserRepository(db)
	hash, _ := bcrypt.GenerateFromPassword([]byte("secretpw"), bcrypt.MinCost)
	if err := repo.Create(ctx, &models.User{
		Username: "loguser", PasswordHash: string(hash), PhoneNumber: "6391712222004",
		Country: "PH", Status: models.UserStatusActive,
	}); err != nil {
		t.Fatal(err)
	}
	smsSvc := newIntegrationSMS(t, db)
	e := echo.New()
	h := NewAuthHandler(repo, smsSvc, nil)
	raw, _ := json.Marshal(map[string]string{"username": "loguser", "password": "secretpw"})
	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewReader(raw))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	if err := h.Login(c); err != nil {
		t.Fatal(err)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("code=%d body=%s", rec.Code, rec.Body.String())
	}
}

func TestLogin_ForbiddenPending(t *testing.T) {
	db := testGormDB(t)
	ctx := context.Background()
	repo := repository.NewUserRepository(db)
	hash, _ := bcrypt.GenerateFromPassword([]byte("pw"), bcrypt.MinCost)
	if err := repo.Create(ctx, &models.User{
		Username: "pend", PasswordHash: string(hash), PhoneNumber: "6391712222005",
		Country: "PH", Status: models.UserStatusPending,
	}); err != nil {
		t.Fatal(err)
	}
	e := echo.New()
	h := NewAuthHandler(repo, &stubSMS{}, nil)
	raw, _ := json.Marshal(map[string]string{"username": "pend", "password": "pw"})
	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewReader(raw))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	if err := h.Login(c); err != nil {
		t.Fatal(err)
	}
	if rec.Code != http.StatusForbidden {
		t.Fatalf("code=%d", rec.Code)
	}
}
