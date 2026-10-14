package handlers

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
)

func TestListMessages_ServiceError(t *testing.T) {
	e := echo.New()
	h := NewSMSHandler(&stubSMS{listErr: errors.New("db")}, nil)
	req := httptest.NewRequest(http.MethodGet, "/messages", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	if err := h.ListMessages(c); err != nil {
		t.Fatal(err)
	}
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("code=%d", rec.Code)
	}
}

func TestGetMessage_NotFound(t *testing.T) {
	e := echo.New()
	h := NewSMSHandler(&stubSMS{notFoundWithGW: true}, nil)
	req := httptest.NewRequest(http.MethodGet, "/messages/x", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("x")
	if err := h.GetMessage(c); err != nil {
		t.Fatal(err)
	}
	if rec.Code != http.StatusNotFound {
		t.Fatalf("code=%d", rec.Code)
	}
}

func TestGetMessage_ServiceError(t *testing.T) {
	e := echo.New()
	h := NewSMSHandler(&stubSMS{gwErr: errors.New("db")}, nil)
	req := httptest.NewRequest(http.MethodGet, "/messages/x", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("x")
	if err := h.GetMessage(c); err != nil {
		t.Fatal(err)
	}
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("code=%d", rec.Code)
	}
}

func TestHandleCallback_ServiceError(t *testing.T) {
	e := echo.New()
	h := NewSMSHandler(&stubSMS{cbErr: errors.New("apply failed")}, nil)
	body := `{"message_id":"p1","status":"Send-success","actual_cost":0}`
	req := httptest.NewRequest(http.MethodPost, "/cb", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	if err := h.HandleCallback(c); err != nil {
		t.Fatal(err)
	}
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("code=%d", rec.Code)
	}
}
