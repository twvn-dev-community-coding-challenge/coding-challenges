package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/dotdak/sms-otp/internal/sms"
	"github.com/labstack/echo/v4"
)

func TestNewSMSHandler(t *testing.T) {
	h := NewSMSHandler(&stubSMS{}, nil)
	if h == nil {
		t.Fatal("nil handler")
	}
}

func TestSendSMS_BindError(t *testing.T) {
	e := echo.New()
	h := NewSMSHandler(&stubSMS{}, sms.NewInMemoryCostTracker())
	req := httptest.NewRequest(http.MethodPost, "/send", strings.NewReader("{"))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	if err := h.SendSMS(c); err != nil {
		t.Fatal(err)
	}
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("code=%d", rec.Code)
	}
}

func TestSendSMS_Validation(t *testing.T) {
	e := echo.New()
	h := NewSMSHandler(&stubSMS{}, sms.NewInMemoryCostTracker())
	req := httptest.NewRequest(http.MethodPost, "/send", strings.NewReader(`{"country":"","phone_number":"","content":""}`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	if err := h.SendSMS(c); err != nil {
		t.Fatal(err)
	}
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("code=%d", rec.Code)
	}
}

func TestSendSMS_Accepted(t *testing.T) {
	e := echo.New()
	h := NewSMSHandler(&stubSMS{}, sms.NewInMemoryCostTracker())
	body := `{"country":"PH","phone_number":"6391712345678","content":"hello"}`
	req := httptest.NewRequest(http.MethodPost, "/send", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	if err := h.SendSMS(c); err != nil {
		t.Fatal(err)
	}
	if rec.Code != http.StatusAccepted {
		t.Fatalf("code=%d", rec.Code)
	}
}

func TestHandleCallback_BindError(t *testing.T) {
	e := echo.New()
	h := NewSMSHandler(&stubSMS{}, nil)
	req := httptest.NewRequest(http.MethodPost, "/cb", strings.NewReader("x"))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	if err := h.HandleCallback(c); err != nil {
		t.Fatal(err)
	}
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("code=%d", rec.Code)
	}
}

func TestHandleCallback_Validate(t *testing.T) {
	e := echo.New()
	h := NewSMSHandler(&stubSMS{}, nil)
	req := httptest.NewRequest(http.MethodPost, "/cb", strings.NewReader(`{"message_id":"","status":""}`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	if err := h.HandleCallback(c); err != nil {
		t.Fatal(err)
	}
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("code=%d", rec.Code)
	}
}

func TestHandleCallback_OK(t *testing.T) {
	e := echo.New()
	h := NewSMSHandler(&stubSMS{}, nil)
	req := httptest.NewRequest(http.MethodPost, "/cb", strings.NewReader(`{"message_id":"p1","status":"Send-success","actual_cost":0}`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	if err := h.HandleCallback(c); err != nil {
		t.Fatal(err)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("code=%d body=%s", rec.Code, rec.Body.String())
	}
}

func TestListMessages_BadSince(t *testing.T) {
	e := echo.New()
	h := NewSMSHandler(&stubSMS{}, nil)
	req := httptest.NewRequest(http.MethodGet, "/messages?since=bad", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	if err := h.ListMessages(c); err != nil {
		t.Fatal(err)
	}
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("code=%d", rec.Code)
	}
}

func TestListMessages_BadLimit(t *testing.T) {
	e := echo.New()
	h := NewSMSHandler(&stubSMS{}, nil)
	req := httptest.NewRequest(http.MethodGet, "/messages?limit=0", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	if err := h.ListMessages(c); err != nil {
		t.Fatal(err)
	}
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("code=%d", rec.Code)
	}
}

func TestListMessages_BadOffset(t *testing.T) {
	e := echo.New()
	h := NewSMSHandler(&stubSMS{}, nil)
	req := httptest.NewRequest(http.MethodGet, "/messages?offset=-1", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	if err := h.ListMessages(c); err != nil {
		t.Fatal(err)
	}
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("code=%d", rec.Code)
	}
}

func TestListMessages_OK(t *testing.T) {
	e := echo.New()
	h := NewSMSHandler(&stubSMS{}, nil)
	req := httptest.NewRequest(http.MethodGet, "/messages", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	if err := h.ListMessages(c); err != nil {
		t.Fatal(err)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("code=%d", rec.Code)
	}
}

func TestGetMessage_MissingID(t *testing.T) {
	e := echo.New()
	h := NewSMSHandler(&stubSMS{}, nil)
	req := httptest.NewRequest(http.MethodGet, "/messages/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("")
	if err := h.GetMessage(c); err != nil {
		t.Fatal(err)
	}
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("code=%d", rec.Code)
	}
}

func TestGetMessage_OK(t *testing.T) {
	e := echo.New()
	h := NewSMSHandler(&stubSMS{}, nil)
	req := httptest.NewRequest(http.MethodGet, "/messages/x1", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("x1")
	if err := h.GetMessage(c); err != nil {
		t.Fatal(err)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("code=%d", rec.Code)
	}
}

func TestGetStats_NoTracker(t *testing.T) {
	e := echo.New()
	h := NewSMSHandler(&stubSMS{}, nil)
	req := httptest.NewRequest(http.MethodGet, "/stats", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	if err := h.GetStats(c); err != nil {
		t.Fatal(err)
	}
	if rec.Code != http.StatusNotImplemented {
		t.Fatalf("code=%d", rec.Code)
	}
}

func TestGetStats_WithTracker(t *testing.T) {
	e := echo.New()
	h := NewSMSHandler(&stubSMS{}, sms.NewInMemoryCostTracker())
	req := httptest.NewRequest(http.MethodGet, "/stats", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	if err := h.GetStats(c); err != nil {
		t.Fatal(err)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("code=%d", rec.Code)
	}
}
