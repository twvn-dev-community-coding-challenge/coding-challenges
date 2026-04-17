package handlers

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"sms-service/internal/models"
	"sms-service/internal/services"
	"time"

	"sms-service/internal/repositories"

	"github.com/gin-gonic/gin"
)

// SMSMessageHandlerDeps holds the dependencies injected into SMSMessageHandler.
type SMSMessageHandlerDeps struct {
	SMSMessageRepo                repositories.SMSMessageRepository
	CountryRepo                   repositories.CountryRepository
	SenderRepo                    repositories.SenderRepository
	CarrierRepo                   repositories.CarrierRepository
	ProviderAgreementRepo         repositories.ProviderAgreementRepository
	ProviderRepo                  repositories.ProviderRepository
	ProviderSelectionDecisionRepo repositories.ProviderSelectionDecisionRepository
}

// SMSMessageHandler groups all SMS-message HTTP handlers.
type SMSMessageHandler struct {
	deps SMSMessageHandlerDeps
}

// NewSMSMessageHandler constructs a handler wired to the given repositories.
func NewSMSMessageHandler(deps SMSMessageHandlerDeps) *SMSMessageHandler {
	return &SMSMessageHandler{deps: deps}
}

// SendSMS handles POST /sms-message/send
func (h *SMSMessageHandler) SendSMS(c *gin.Context) {
	log.Println("[SendSMS] Received request")
	var req SendSMSRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorEnvelope{
			Error: ErrorBody{Code: "BAD_REQUEST", Message: err.Error()},
			Meta:  Meta{RequestID: fmt.Sprintf("req-%d", time.Now().UnixNano()), Timestamp: time.Now().UTC()},
		})
		return
	}

	ctx := c.Request.Context()

	log.Printf("[SendSMS] Parsed request: sender=%s, phone=%s, country=%s", req.SenderID, req.RecipientPhone, req.CountryCode)

	country, err := h.deps.CountryRepo.GetCountryByCode(req.CountryCode)
	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, ErrorEnvelope{
			Error: ErrorBody{Code: "COUNTRY_NOT_FOUND", Message: "country not found for given country code"},
			Meta:  Meta{RequestID: fmt.Sprintf("req-%d", time.Now().UnixNano()), Timestamp: time.Now().UTC()},
		})
		return
	}

	log.Printf("[SendSMS] Country resolved: id=%s, name=%s", country.ID, country.Name)

	carrier, err := h.deps.CarrierRepo.FindByCountryAndPhone(ctx, country.ID, req.RecipientPhone)
	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, ErrorEnvelope{
			Error: ErrorBody{Code: "CARRIER_NOT_FOUND", Message: "no carrier found for recipient phone"},
			Meta:  Meta{RequestID: fmt.Sprintf("req-%d", time.Now().UnixNano()), Timestamp: time.Now().UTC()},
		})
		return
	}
	log.Printf("[SendSMS] Carrier resolved: id=%s, name=%s", carrier.ID, carrier.Name)

	msg := models.SMSMessage{
		ID:             fmt.Sprintf("msg-%d", time.Now().UnixNano()),
		SenderID:       req.SenderID,
		RecipientPhone: req.RecipientPhone,
		Content:        req.Content,
		CarrierID:      carrier.ID,
		ProviderID:     nil,
		Status:         models.StatusNew,
		CreatedAt:      time.Now().UTC(),
		UpdatedAt:      time.Now().UTC(),
	}

	if err := h.deps.SMSMessageRepo.Save(ctx, msg); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorEnvelope{
			Error: ErrorBody{Code: "INTERNAL_ERROR", Message: "failed to save SMS message"},
			Meta:  Meta{RequestID: fmt.Sprintf("req-%d", time.Now().UnixNano()), Timestamp: time.Now().UTC()},
		})
		return
	}

	log.Printf("[SendSMS] Message created: id=%s, selecting provider...", msg.ID)

	selector := services.NewGetFirstProviderSelector(h.deps.ProviderAgreementRepo, h.deps.ProviderRepo)
	decision, err := selector.Select(&msg)
	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, ErrorEnvelope{
			Error: ErrorBody{Code: "NO_PROVIDER_AGREEMENT", Message: err.Error()},
			Meta:  Meta{RequestID: fmt.Sprintf("req-%d", time.Now().UnixNano()), Timestamp: time.Now().UTC()},
		})
		return
	}

	if h.deps.ProviderSelectionDecisionRepo != nil {
		if err := h.deps.ProviderSelectionDecisionRepo.Save(ctx, decision); err != nil {
			c.JSON(http.StatusInternalServerError, ErrorEnvelope{
				Error: ErrorBody{Code: "INTERNAL_ERROR", Message: "failed to save provider selection decision"},
				Meta:  Meta{RequestID: fmt.Sprintf("req-%d", time.Now().UnixNano()), Timestamp: time.Now().UTC()},
			})
			return
		}
	}

	log.Printf("[SendSMS] Provider selected: provider=%s, estimated_cost=%s %s", decision.ProviderID, decision.EstimatedCost.Amount, decision.EstimatedCost.Currency)

	msg.ProviderID = &decision.ProviderID
	msg.EstimatedCost = decision.EstimatedCost
	msg.Status = models.StatusSendToProvider
	msg.UpdatedAt = time.Now().UTC()

	if err := h.deps.SMSMessageRepo.Save(ctx, msg); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorEnvelope{
			Error: ErrorBody{Code: "INTERNAL_ERROR", Message: "failed to update SMS message"},
			Meta:  Meta{RequestID: fmt.Sprintf("req-%d", time.Now().UnixNano()), Timestamp: time.Now().UTC()},
		})
		return
	}

	meta := Meta{
		RequestID: fmt.Sprintf("req-%d", time.Now().UnixNano()),
		Timestamp: time.Now().UTC(),
	}
	log.Printf("[SendSMS] Success: id=%s, status=%s, provider=%s", msg.ID, msg.Status, decision.ProviderID)

	c.JSON(http.StatusCreated, SendSMSResponseEnvelope{
		Data: SendSMSResponse{
			MessageID:     msg.ID,
			Status:        string(msg.Status),
			ProviderID:    decision.ProviderID,
			CarrierID:     msg.CarrierID,
			EstimatedCost: msg.EstimatedCost.Amount,
			Currency:      msg.EstimatedCost.Currency,
			CreatedAt:     msg.CreatedAt,
		},
		Meta: meta,
	})
}

// HandleProviderCallback handles POST /sms-message/webhooks/provider-callback
func (h *SMSMessageHandler) HandleProviderCallback(c *gin.Context) {
	log.Println("[Callback] Received webhook")
	ctx := c.Request.Context()

	var req ProviderCallbackRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorEnvelope{
			Error: ErrorBody{Code: "BAD_REQUEST", Message: "invalid request body"},
			Meta:  Meta{RequestID: fmt.Sprintf("req-%d", time.Now().UnixNano()), Timestamp: time.Now().UTC()},
		})
		return
	}

	messageID := req.MessageID
	if messageID == "" {
		c.JSON(http.StatusBadRequest, ErrorEnvelope{
			Error: ErrorBody{Code: "BAD_REQUEST", Message: "missing message_id"},
			Meta:  Meta{RequestID: fmt.Sprintf("req-%d", time.Now().UnixNano()), Timestamp: time.Now().UTC()},
		})
		return
	}

	log.Printf("[Callback] Processing: message_id=%s, status=%s, event=%s", messageID, req.Status, req.Event)

	message, err := h.deps.SMSMessageRepo.GetById(ctx, messageID)
	if err != nil {
		if errors.Is(err, models.ErrNotFound) {
			c.JSON(http.StatusNotFound, ErrorEnvelope{
				Error: ErrorBody{Code: "NOT_FOUND", Message: fmt.Sprintf("SMS message %q not found", messageID)},
				Meta:  Meta{RequestID: fmt.Sprintf("req-%d", time.Now().UnixNano()), Timestamp: time.Now().UTC()},
			})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorEnvelope{
			Error: ErrorBody{Code: "INTERNAL_ERROR", Message: "failed to retrieve SMS message"},
			Meta:  Meta{RequestID: fmt.Sprintf("req-%d", time.Now().UnixNano()), Timestamp: time.Now().UTC()},
		})
		return
	}

	if !message.CanTransitionTo(models.SMSStatus(req.Status)) {
		log.Printf("Failed to transition %s: current status=%s, target status=%s", message.ID, message.Status, req.Status)
		c.JSON(http.StatusBadRequest, ErrorEnvelope{
			Error: ErrorBody{Code: "INVALID_STATUS", Message: fmt.Sprintf("invalid status transition from %s to %s", message.Status, req.Status)},
			Meta:  Meta{RequestID: fmt.Sprintf("req-%d", time.Now().UnixNano()), Timestamp: time.Now().UTC()},
		})
		return
	}
	// Assuming the event name is the same as the status
	newStatus := models.SMSStatus(req.Status)
	if newStatus == "" {
		c.JSON(http.StatusUnprocessableEntity, ErrorEnvelope{
			Error: ErrorBody{Code: "INVALID_STATUS", Message: fmt.Sprintf("unknown status %q", req.Status)},
			Meta:  Meta{RequestID: fmt.Sprintf("req-%d", time.Now().UnixNano()), Timestamp: time.Now().UTC()},
		})
		return
	}

	switch newStatus {
	case models.StatusSendFailed, models.StatusCarrierRejected:
		message.Status = newStatus
		failureReason := req.FailureReason
		if failureReason != "" {
			message.FailureReason = failureReason
		} else {
			failureReason = "Unknown failure reason"
			message.FailureReason = failureReason
		}
	case models.StatusSendSuccess:
		message.Status = newStatus

		if req.ActualCost != "" {
			money := models.Money{
				Amount:   req.ActualCost,
				Currency: "VND",
			}
			message.ActualCost = &money
		}
	default:
		message.Status = newStatus
	}

	log.Printf("[Callback] Updating message %s: status=%s", message.ID, message.Status)

	if err := h.deps.SMSMessageRepo.Save(ctx, message); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorEnvelope{
			Error: ErrorBody{Code: "INTERNAL_ERROR", Message: "failed to save SMS message callback"},
			Meta:  Meta{RequestID: fmt.Sprintf("req-%d", time.Now().UnixNano()), Timestamp: time.Now().UTC()},
		})
		return
	}

	log.Printf("[Callback] Success: message_id=%s, new_status=%s", message.ID, message.Status)

	c.JSON(http.StatusOK, ProviderCallbackResponseEnvelope{
		Data: ProviderCallbackResponse{
			MessageID: message.ID,
			Status:    string(message.Status),
			UpdatedAt: message.UpdatedAt,
		},
		Meta: Meta{
			RequestID: fmt.Sprintf("req-%d", time.Now().UnixNano()),
			Timestamp: time.Now().UTC(),
		},
	})
}

// GetByID handles GET /sms-message/:id
func (h *SMSMessageHandler) GetByID(c *gin.Context) {
	id := c.Param("id")
	log.Printf("[GetByID] Fetching message: id=%s", id)
	ctx := c.Request.Context()

	meta := Meta{
		RequestID: fmt.Sprintf("req-%d", time.Now().UnixNano()),
		Timestamp: time.Now().UTC(),
	}

	msg, err := h.deps.SMSMessageRepo.GetById(ctx, id)
	if err != nil {
		if errors.Is(err, models.ErrNotFound) {
			c.JSON(http.StatusNotFound, ErrorEnvelope{
				Error: ErrorBody{
					Code:    "NOT_FOUND",
					Message: fmt.Sprintf("SMS message %q not found", id),
				},
				Meta: meta,
			})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorEnvelope{
			Error: ErrorBody{
				Code:    "INTERNAL_ERROR",
				Message: "unexpected error retrieving message",
			},
			Meta: meta,
		})
		return
	}

	msgDTO := SMSMessageDTO{
		ID:            msg.ID,
		SenderID:      msg.SenderID,
		RecipientID:   msg.RecipientID,
		Content:       msg.Content,
		Status:        string(msg.Status),
		EstimatedCost: msg.EstimatedCost.Amount,
		Currency:      msg.EstimatedCost.Currency,
		ProviderID: func() string {
			if msg.ProviderID != nil {
				return *msg.ProviderID
			}
			return ""
		}(),
		CarrierID:         msg.CarrierID,
		ProviderMessageID: msg.ProviderMessageID,
		FailureReason:     msg.FailureReason,
		CreatedAt:         msg.CreatedAt,
		UpdatedAt:         msg.UpdatedAt,
	}
	if msg.ActualCost != nil {
		msgDTO.ActualCost = msg.ActualCost.Amount
	}

	var decisionDTO *ProviderSelectionDecisionDTO

	log.Printf("[GetByID] Found message: id=%s, status=%s", msg.ID, msg.Status)

	c.JSON(http.StatusOK, GetSMSMessageResponseEnvelope{
		Data: GetSMSMessageResponse{
			Message:  msgDTO,
			Decision: decisionDTO,
		},
		Meta: meta,
	})
}
