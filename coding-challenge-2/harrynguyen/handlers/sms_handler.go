package handlers

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/dotdak/sms-otp/internal/obslog"
	"github.com/dotdak/sms-otp/internal/providers"
	"github.com/dotdak/sms-otp/internal/sms"
	"github.com/labstack/echo/v4"
)

type SMSHandler struct {
	smsService  sms.SMSCoreService
	costTracker *sms.InMemoryCostTracker
}

func NewSMSHandler(smsService sms.SMSCoreService, costTracker *sms.InMemoryCostTracker) *SMSHandler {
	return &SMSHandler{
		smsService:  smsService,
		costTracker: costTracker,
	}
}

type SendSMSRequest struct {
	Country     string `json:"country"`
	PhoneNumber string `json:"phone_number"`
	Content     string `json:"content"`
}

func (h *SMSHandler) SendSMS(c echo.Context) error {
	var req SendSMSRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request payload"})
	}

	if req.Country == "" || req.PhoneNumber == "" || req.Content == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Country, phone number, and content are required"})
	}

	ctx := sms.WithSendSource(sms.WithTestSMSMode(c.Request().Context(), testSMSModeFromRequest(c)), sms.SendSourceAPI)
	msg, err := h.smsService.SendSMS(ctx, req.Country, req.PhoneNumber, req.Content)
	if err != nil {
		return respondSendSMSError(c, err)
	}

	return c.JSON(http.StatusAccepted, map[string]interface{}{
		"message":        "SMS queued for sending",
		"id":             msg.ID,
		"provider_id":    msg.MessageID,
		"status":         msg.Status,
		"carrier":        msg.Carrier,
		"provider":       msg.Provider,
		"estimated_cost": msg.EstimatedCost,
	})
}

type CallbackRequest struct {
	MessageID  string                  `json:"message_id"`
	Status     providers.MessageStatus `json:"status"`
	ActualCost float64                 `json:"actual_cost"`
}

func (h *SMSHandler) HandleCallback(c echo.Context) error {
	ctx := c.Request().Context()
	obslog.Init()

	var req CallbackRequest
	if err := c.Bind(&req); err != nil {
		obslog.L.WarnContext(ctx, obslog.MsgSMSCallbackBindFailed, append(smsCallbackAttrs(ctx, obslog.StepBind), "err", err)...)
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request payload"})
	}

	if req.MessageID == "" || req.Status == "" {
		obslog.L.WarnContext(ctx, obslog.MsgSMSCallbackValidateFailed, append(smsCallbackAttrs(ctx, obslog.StepValidate),
			"has_message_id", req.MessageID != "",
			"has_status", req.Status != "",
		)...)
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Message ID and Status are required"})
	}

	if err := h.smsService.HandleCallback(ctx, req.MessageID, req.Status, req.ActualCost); err != nil {
		// DefaultSMSService logs lookup/apply failures with full context; respond without duplicating error logs.
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": fmt.Sprintf("Failed to handle callback: %v", err)})
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "Callback processed successfully"})
}

func (h *SMSHandler) ListMessages(c echo.Context) error {
	params := sms.MessageListParams{}
	if s := c.QueryParam("status"); s != "" {
		params.Status = providers.MessageStatus(s)
	}
	params.Phone = c.QueryParam("phone")
	if sinceStr := c.QueryParam("since"); sinceStr != "" {
		t, err := time.Parse(time.RFC3339, sinceStr)
		if err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "since must be RFC3339 datetime"})
		}
		params.Since = &t
	}
	limit := 100
	if v := c.QueryParam("limit"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n < 1 {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "limit must be a positive integer"})
		}
		if n > 500 {
			n = 500
		}
		limit = n
	}
	params.Limit = limit
	if v := c.QueryParam("offset"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n < 0 {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "offset must be a non-negative integer"})
		}
		params.Offset = n
	}

	msgs, err := h.smsService.ListMessages(c.Request().Context(), params)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": fmt.Sprintf("Failed to list messages: %v", err)})
	}

	return c.JSON(http.StatusOK, msgs)
}

// GetMessage returns one SMS row and its status log timeline (newest-first list ordering applies only to list; logs are chronological).
func (h *SMSHandler) GetMessage(c echo.Context) error {
	id := c.Param("id")
	if id == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "id is required"})
	}
	msg, logs, err := h.smsService.GetMessageWithLogs(c.Request().Context(), id)
	if err != nil {
		if errors.Is(err, sms.ErrNotFound) {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "message not found"})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": fmt.Sprintf("Failed to load message: %v", err)})
	}
	return c.JSON(http.StatusOK, map[string]interface{}{
		"message":     msg,
		"status_logs": logs,
	})
}

func (h *SMSHandler) GetStats(c echo.Context) error {
	if h.costTracker == nullTracker() {
		return c.JSON(http.StatusNotImplemented, map[string]string{"error": "Cost tracking is not enabled"})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"by_provider": h.costTracker.GetProviderMetrics(),
		"by_country":  h.costTracker.GetCountryMetrics(),
	})
}

// helper func for checking nil tracker safety
func nullTracker() *sms.InMemoryCostTracker {
	return nil
}
