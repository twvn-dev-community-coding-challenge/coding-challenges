package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/dotdak/sms-otp/internal/carrier"
	"github.com/dotdak/sms-otp/internal/providers"
	"github.com/dotdak/sms-otp/internal/sms"
	"github.com/dotdak/sms-otp/models"
	"github.com/dotdak/sms-otp/repository"
	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func TestRegister_UsernameConflict(t *testing.T) {
	db := testGormDB(t)
	ctx := context.Background()
	userRepo := repository.NewUserRepository(db)
	hash, _ := bcrypt.GenerateFromPassword([]byte("pw"), bcrypt.MinCost)
	if err := userRepo.Create(ctx, &models.User{
		Username: "dup", PasswordHash: string(hash), PhoneNumber: "6391711111001",
		Country: "PH", Status: models.UserStatusPending,
	}); err != nil {
		t.Fatal(err)
	}

	smsSvc := newIntegrationSMS(t, db)
	h := NewAuthHandler(userRepo, smsSvc, nil)
	e := echo.New()
	body := `{"username":"dup","password":"pw123456","phone_number":"6391711111002","country":"PH"}`
	req := httptest.NewRequest(http.MethodPost, "/register", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	if err := h.Register(c); err != nil {
		t.Fatal(err)
	}
	if rec.Code != http.StatusConflict {
		t.Fatalf("code=%d", rec.Code)
	}
}

func TestRegister_PhoneConflict(t *testing.T) {
	db := testGormDB(t)
	ctx := context.Background()
	userRepo := repository.NewUserRepository(db)
	hash, _ := bcrypt.GenerateFromPassword([]byte("pw"), bcrypt.MinCost)
	if err := userRepo.Create(ctx, &models.User{
		Username: "u1", PasswordHash: string(hash), PhoneNumber: "6391711111003",
		Country: "PH", Status: models.UserStatusPending,
	}); err != nil {
		t.Fatal(err)
	}

	smsSvc := newIntegrationSMS(t, db)
	h := NewAuthHandler(userRepo, smsSvc, nil)
	e := echo.New()
	body := `{"username":"u2","password":"pw123456","phone_number":"6391711111003","country":"PH"}`
	req := httptest.NewRequest(http.MethodPost, "/register", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	if err := h.Register(c); err != nil {
		t.Fatal(err)
	}
	if rec.Code != http.StatusConflict {
		t.Fatalf("code=%d", rec.Code)
	}
}

func newIntegrationSMS(t *testing.T, db *gorm.DB) sms.SMSCoreService {
	t.Helper()
	smsRepo := repository.NewSMSGormRepository(db)
	resolver := carrier.NewPrefixCarrierResolver()
	router := providers.NewSimpleProviderRouter()
	router.RegisterRouter("PH", providers.CarrierGlobe, providers.ProviderMessageBird, &providers.MockProvider{Name: providers.ProviderMessageBird, SkipDelay: true, Quiet: true})
	return sms.NewSMSServiceWithImmediateDispatch(smsRepo, resolver, router, nil)
}
