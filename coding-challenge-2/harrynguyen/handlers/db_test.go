package handlers

import (
	"testing"

	"github.com/dotdak/sms-otp/models"
	"github.com/dotdak/sms-otp/repository"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func testGormDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file:handlers_test?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&models.User{}); err != nil {
		t.Fatal(err)
	}
	if err := repository.AutoMigrateSMS(db); err != nil {
		t.Fatal(err)
	}
	return db
}
