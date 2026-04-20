package handlers

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
)

func TestSendSMS_ServiceError(t *testing.T) {
	e := echo.New()
	h := NewSMSHandler(&stubSMS{sendErr: errors.New("queue down")}, nil)
	body := `{"country":"PH","phone_number":"6391712345678","content":"hello"}`
	req := httptest.NewRequest(http.MethodPost, "/send", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	if err := h.SendSMS(c); err != nil {
		t.Fatal(err)
	}
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("code=%d body=%s", rec.Code, rec.Body.String())
	}
}
