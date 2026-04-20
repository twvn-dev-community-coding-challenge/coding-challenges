package repository

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/dotdak/sms-otp/internal/providers"
	"github.com/dotdak/sms-otp/internal/sms"
	"github.com/dotdak/sms-otp/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func newTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file:repo_test?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&models.User{}); err != nil {
		t.Fatal(err)
	}
	if err := AutoMigrateSMS(db); err != nil {
		t.Fatal(err)
	}
	return db
}

func TestUserRepository_CRUD(t *testing.T) {
	db := newTestDB(t)
	repo := NewUserRepository(db)
	ctx := context.Background()

	u := &models.User{
		Username:     "alice",
		PasswordHash: "hash",
		PhoneNumber:  "841234567890",
		Country:      "VN",
		Status:       models.UserStatusPending,
	}
	if err := repo.Create(ctx, u); err != nil {
		t.Fatal(err)
	}
	if u.ID == 0 {
		t.Fatal("expected user id")
	}

	got, err := repo.GetByUsername(ctx, "alice")
	if err != nil || got == nil || got.Username != "alice" {
		t.Fatalf("GetByUsername: err=%v user=%v", err, got)
	}

	got2, err := repo.GetByPhoneNumber(ctx, "841234567890")
	if err != nil || got2 == nil {
		t.Fatalf("GetByPhoneNumber: err=%v user=%v", err, got2)
	}

	got3, err := repo.GetByID(ctx, u.ID)
	if err != nil || got3 == nil {
		t.Fatalf("GetByID: err=%v user=%v", err, got3)
	}

	u.Status = models.UserStatusActive
	if err := repo.Update(ctx, u); err != nil {
		t.Fatal(err)
	}

	if err := repo.Delete(ctx, u.ID); err != nil {
		t.Fatal(err)
	}
	nilUser, err := repo.GetByUsername(ctx, "alice")
	if err != nil || nilUser != nil {
		t.Fatalf("after delete want nil user, got %v err=%v", nilUser, err)
	}
}

func TestUserRepository_NotFound(t *testing.T) {
	db := newTestDB(t)
	repo := NewUserRepository(db)
	ctx := context.Background()

	u, err := repo.GetByUsername(ctx, "nobody")
	if err != nil || u != nil {
		t.Fatalf("want nil user, got %v err=%v", u, err)
	}
}

func TestSMSGormRepository_CRUD(t *testing.T) {
	db := newTestDB(t)
	repo := NewSMSGormRepository(db)
	ctx := context.Background()

	msg := &providers.SMSMessage{
		SendSource: "api",
		Country:    "PH",
		PhoneNumber: "6391712345678",
		Content:    "hi",
		Carrier:    providers.CarrierGlobe,
		Provider:   providers.ProviderMessageBird,
		Status:     providers.StatusNew,
	}
	if err := repo.Create(ctx, msg); err != nil {
		t.Fatal(err)
	}
	if msg.ID == "" {
		t.Fatal("expected id")
	}

	got, err := repo.GetByID(ctx, msg.ID)
	if err != nil || got == nil || got.PhoneNumber != msg.PhoneNumber {
		t.Fatalf("GetByID: %v %v", err, got)
	}

	msg.MessageID = "prov-123"
	msg.Status = providers.StatusQueue
	if err := repo.Update(ctx, msg); err != nil {
		t.Fatal(err)
	}

	byProv, err := repo.GetByMessageID(ctx, "prov-123")
	if err != nil || byProv == nil || byProv.ID != msg.ID {
		t.Fatalf("GetByMessageID: err=%v msg=%v", err, byProv)
	}

	since := time.Now().Add(-time.Hour)
	list, err := repo.ListMessages(ctx, sms.MessageListParams{
		Status: providers.StatusQueue,
		Phone:  "1712345",
		Since:  &since,
		Limit:  10,
		Offset: 0,
	})
	if err != nil || len(list) != 1 {
		t.Fatalf("ListMessages: err=%v n=%d", err, len(list))
	}

	log := &providers.StatusLog{SMSID: msg.ID, Status: providers.StatusNew, Metadata: "x"}
	if err := repo.AddStatusLog(ctx, log); err != nil {
		t.Fatal(err)
	}
	logs, err := repo.GetStatusLogs(ctx, msg.ID)
	if err != nil || len(logs) != 1 {
		t.Fatalf("GetStatusLogs: err=%v len=%d", err, len(logs))
	}

	err = repo.Update(ctx, &providers.SMSMessage{ID: "missing", Status: providers.StatusNew})
	if err == nil || !errors.Is(err, sms.ErrNotFound) {
		t.Fatalf("expected ErrNotFound on update missing row, got %v", err)
	}
}
