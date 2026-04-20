package handlers

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/dotdak/sms-otp/internal/httputil"
	"github.com/dotdak/sms-otp/internal/obslog"
	"github.com/dotdak/sms-otp/internal/phone"
	"github.com/dotdak/sms-otp/internal/providers"
	"github.com/dotdak/sms-otp/internal/ratelimit"
	"github.com/dotdak/sms-otp/internal/sms"
	"github.com/dotdak/sms-otp/models"
	"github.com/dotdak/sms-otp/repository"
	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"
)

const defaultOTPTTL = 5 * time.Minute

func getOTPTTL() time.Duration {
	ttlEnv := os.Getenv("OTP_TTL_SECONDS")
	if ttlEnv == "" {
		return defaultOTPTTL
	}

	secs, err := strconv.Atoi(ttlEnv)
	if err != nil || secs <= 0 {
		return defaultOTPTTL
	}

	return time.Duration(secs) * time.Second
}

const (
	regStartsPerIPPerHour     = 5
	regStartsPerDevicePerHour = 5
	loginAttemptsPerIPPerHour = 20
	loginVerifyPerIPPerHour   = 40
)

type authSMSObserver struct{}

func (authSMSObserver) OnMessageUpdated(_ context.Context, _ *providers.SMSMessage) {}

type AuthHandler struct {
	userRepo   repository.UserRepository
	smsService sms.SMSCoreService
	limiter    *ratelimit.Limiter
}

func NewAuthHandler(userRepo repository.UserRepository, smsService sms.SMSCoreService, limiter *ratelimit.Limiter) *AuthHandler {
	if smsService != nil {
		smsService.RegisterSourceObserver(sms.SendSourceAuth, authSMSObserver{})
	}
	return &AuthHandler{
		userRepo:   userRepo,
		smsService: smsService,
		limiter:    limiter,
	}
}

type RegisterRequest struct {
	Username      string `json:"username"`
	Password      string `json:"password"`
	PhoneNumber   string `json:"phone_number"`
	Country       string `json:"country"`
	OTPTTLSeconds *int   `json:"otp_ttl_seconds,omitempty"`
}

func (h *AuthHandler) Register(c echo.Context) error {
	ctx := c.Request().Context()
	var req RegisterRequest
	if err := c.Bind(&req); err != nil {
		obslog.Init()
		obslog.L.ErrorContext(ctx, obslog.MsgRegisterFailed,
			append(authFlowAttrs(ctx, obslog.FlowRegister, obslog.StepParseBody), "err", err.Error())...,
		)
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request payload"})
	}

	if req.Username == "" || req.Password == "" || req.PhoneNumber == "" || req.Country == "" {
		obslog.Init()
		obslog.L.WarnContext(ctx, obslog.MsgRegisterRejected,
			append(authFlowAttrs(ctx, obslog.FlowRegister, obslog.StepValidateRequired),
				"username_empty", req.Username == "",
				"password_empty", req.Password == "",
				"phone_empty", req.PhoneNumber == "",
				"country_empty", req.Country == "",
			)...,
		)
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Username, password, phone number and country are required"})
	}

	iso, err := phone.NormalizeCountry(req.Country)
	if err != nil {
		obslog.Init()
		obslog.L.WarnContext(ctx, obslog.MsgRegisterRejected,
			append(authFlowAttrs(ctx, obslog.FlowRegister, obslog.StepNormalizeCountry),
				"country_input", req.Country,
				"err", err.Error(),
			)...,
		)
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid or unsupported country (use ISO code: VN, TH, SG, PH)"})
	}
	canonicalPhone, err := phone.Validate(iso, req.PhoneNumber)
	if err != nil {
		obslog.Init()
		obslog.L.WarnContext(ctx, obslog.MsgRegisterRejected,
			append(authFlowAttrs(ctx, obslog.FlowRegister, obslog.StepValidatePhone),
				"country_iso", iso,
				"phone_masked", maskPhoneSuffix(phone.CanonicalDigits(req.PhoneNumber)),
				"err", err.Error(),
			)...,
		)
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	if h.limiter != nil {
		hour := time.Now().Unix() / 3600
		ipKey := fmt.Sprintf("reg:1h:ip:%s:%d", httputil.ClientIP(c), hour)
		if err := h.limiter.Allow(ctx, ipKey, regStartsPerIPPerHour, time.Hour); err != nil {
			obslog.Init()
			obslog.L.WarnContext(ctx, obslog.MsgRegisterRejected,
				append(authFlowAttrs(ctx, obslog.FlowRegister, obslog.StepRateLimitIP),
					"limiter_key", ipKey,
					"err", err.Error(),
				)...,
			)
			return respondTooManyRequests(c, err)
		}
		if fp := c.Request().Header.Get("X-Device-Fingerprint"); fp != "" {
			devKey := fmt.Sprintf("reg:1h:dev:%s:%d", fp, hour)
			if err := h.limiter.Allow(ctx, devKey, regStartsPerDevicePerHour, time.Hour); err != nil {
				obslog.Init()
				obslog.L.WarnContext(ctx, obslog.MsgRegisterRejected,
					append(authFlowAttrs(ctx, obslog.FlowRegister, obslog.StepRateLimitDevice),
						"limiter_key", devKey,
						"err", err.Error(),
					)...,
				)
				return respondTooManyRequests(c, err)
			}
		}
	}

	// Check if username already exists
	existingUsername, err := h.userRepo.GetByUsername(ctx, req.Username)
	if err != nil {
		obslog.Init()
		obslog.L.ErrorContext(ctx, obslog.MsgRegisterFailed,
			append(authFlowAttrs(ctx, obslog.FlowRegister, obslog.StepDBLookupUsername),
				"username", req.Username,
				"err", err.Error(),
			)...,
		)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Database error"})
	}
	if existingUsername != nil {
		obslog.Init()
		obslog.L.WarnContext(ctx, obslog.MsgRegisterRejected,
			append(authFlowAttrs(ctx, obslog.FlowRegister, obslog.StepConflictUsername),
				"username", req.Username,
			)...,
		)
		return c.JSON(http.StatusConflict, map[string]string{"error": "Username already taken"})
	}

	// Check if phone number already exists
	existingPhone, err := h.userRepo.GetByPhoneNumber(ctx, canonicalPhone)
	if err != nil {
		obslog.Init()
		obslog.L.ErrorContext(ctx, obslog.MsgRegisterFailed,
			append(authFlowAttrs(ctx, obslog.FlowRegister, obslog.StepDBLookupPhone),
				"phone_masked", maskPhoneSuffix(canonicalPhone),
				"err", err.Error(),
			)...,
		)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Database error"})
	}
	if existingPhone != nil {
		obslog.Init()
		obslog.L.WarnContext(ctx, obslog.MsgRegisterRejected,
			append(authFlowAttrs(ctx, obslog.FlowRegister, obslog.StepConflictPhone),
				"phone_masked", maskPhoneSuffix(canonicalPhone),
			)...,
		)
		return c.JSON(http.StatusConflict, map[string]string{"error": "Phone number already taken"})
	}

	// Hash the password
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		obslog.Init()
		obslog.L.ErrorContext(ctx, obslog.MsgRegisterFailed,
			append(authFlowAttrs(ctx, obslog.FlowRegister, obslog.StepHashPassword), "err", err.Error())...,
		)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to hash password"})
	}

	// Generate 6-digit OTP
	rand.Seed(time.Now().UnixNano())
	otpCode := fmt.Sprintf("%06d", rand.Intn(1000000))
	otpTTL := getOTPTTL()
	if req.OTPTTLSeconds != nil && *req.OTPTTLSeconds > 0 && *req.OTPTTLSeconds <= 3600 {
		otpTTL = time.Duration(*req.OTPTTLSeconds) * time.Second
	}
	expiresAt := time.Now().Add(otpTTL)

	// Create the user in pending status
	user := &models.User{
		Username:     req.Username,
		PasswordHash: string(hashedBytes),
		PhoneNumber:  canonicalPhone,
		Country:      iso,
		OTPCode:      otpCode,
		OTPExpiresAt: &expiresAt,
		Status:       models.UserStatusPending,
	}

	if err := h.userRepo.Create(ctx, user); err != nil {
		obslog.Init()
		obslog.L.ErrorContext(ctx, obslog.MsgRegisterFailed,
			append(authFlowAttrs(ctx, obslog.FlowRegister, obslog.StepDBCreateUser),
				"username", req.Username,
				"phone_masked", maskPhoneSuffix(canonicalPhone),
				"err", err.Error(),
			)...,
		)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create user"})
	}

	// Send OTP via SMS
	smsContent := fmt.Sprintf("Your registration OTP is: %s", otpCode)
	smsCtx := sms.WithSendSource(sms.WithTestSMSMode(ctx, testSMSModeFromRequest(c)), sms.SendSourceAuth)
	smsMsg, err := h.smsService.SendSMS(smsCtx, req.Country, canonicalPhone, smsContent)
	if err != nil {
		obslog.Init()
		obslog.L.ErrorContext(ctx, obslog.MsgRegisterFailed,
			append(authFlowAttrs(ctx, obslog.FlowRegister, obslog.StepSMSSendOTP),
				"user_id", user.ID,
				"username", req.Username,
				"country", req.Country,
				"phone_masked", maskPhoneSuffix(canonicalPhone),
				"err", err.Error(),
			)...,
		)
		if delErr := h.userRepo.Delete(ctx, user.ID); delErr != nil {
			obslog.L.ErrorContext(ctx, obslog.MsgRegisterRollbackFailed,
				append(authFlowAttrs(ctx, obslog.FlowRegister, obslog.StepRollbackDeleteUser),
					"user_id", user.ID,
					"err", delErr.Error(),
				)...,
			)
		} else {
			obslog.L.WarnContext(ctx, obslog.MsgRegisterRolledBack,
				append(authFlowAttrs(ctx, obslog.FlowRegister, obslog.StepRollbackDeleteUser),
					"user_id", user.ID,
					"reason", obslog.ReasonSMSSendFailed,
				)...,
			)
		}
		var lim *ratelimit.LimitedError
		if errors.As(err, &lim) {
			return respondTooManyRequests(c, err)
		}
		if errors.Is(err, phone.ErrInvalidCountry) || errors.Is(err, phone.ErrInvalidNumber) {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": fmt.Sprintf("Failed to send SMS: %v", err)})
	}

	obslog.Init()
	obslog.L.InfoContext(ctx, obslog.MsgRegisterSMSSent,
		append(authFlowAttrs(ctx, obslog.FlowRegister, obslog.StepSMSSent),
			"user_id", user.ID,
			"username", req.Username,
			"sms_id", smsMsg.ID,
			"provider_message_id", smsMsg.MessageID,
			"phone_masked", maskPhoneSuffix(canonicalPhone),
		)...,
	)

	if !strings.EqualFold(os.Getenv("LOG_REGISTRATION_OTP"), "false") {
		obslog.L.InfoContext(ctx, obslog.MsgRegistrationOTP,
			"username", user.Username,
			"otp", otpCode,
			"request_id", obslog.RequestID(ctx),
			"correlation_id", obslog.CorrelationID(ctx),
			"sms_id", smsMsg.ID,
			"provider_message_id", smsMsg.MessageID,
		)
	}

	return c.JSON(http.StatusCreated, map[string]interface{}{
		"message": "User registered successfully. Please verify your phone number with the OTP sent.",
		"user": map[string]interface{}{
			"id":           user.ID,
			"username":     user.Username,
			"phone_number": user.PhoneNumber,
			"status":       user.Status,
		},
	})
}

type VerifyRequest struct {
	Username string `json:"username"`
	OTPCode  string `json:"otp_code"`
}

func (h *AuthHandler) VerifyOTP(c echo.Context) error {
	ctx := c.Request().Context()
	var req VerifyRequest
	if err := c.Bind(&req); err != nil {
		obslog.Init()
		obslog.L.ErrorContext(ctx, obslog.MsgVerifyOTPFailed,
			append(authFlowAttrs(ctx, obslog.FlowVerifyOTP, obslog.StepParseBody), "err", err.Error())...,
		)
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request payload"})
	}

	if req.Username == "" || req.OTPCode == "" {
		obslog.Init()
		obslog.L.WarnContext(ctx, obslog.MsgVerifyOTPRejected,
			append(authFlowAttrs(ctx, obslog.FlowVerifyOTP, obslog.StepValidateRequired),
				"username_empty", req.Username == "",
				"otp_empty", req.OTPCode == "",
			)...,
		)
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Username and OTP code are required"})
	}

	user, err := h.userRepo.GetByUsername(ctx, req.Username)
	if err != nil {
		obslog.Init()
		obslog.L.ErrorContext(ctx, obslog.MsgVerifyOTPFailed,
			append(authFlowAttrs(ctx, obslog.FlowVerifyOTP, obslog.StepDBLookupUsername),
				"username", req.Username,
				"err", err.Error(),
			)...,
		)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Database error"})
	}
	if user == nil {
		obslog.Init()
		obslog.L.WarnContext(ctx, obslog.MsgVerifyOTPRejected,
			append(authFlowAttrs(ctx, obslog.FlowVerifyOTP, obslog.StepUserNotFound),
				"username", req.Username,
			)...,
		)
		return c.JSON(http.StatusNotFound, map[string]string{"error": "User not found"})
	}

	if user.Status == models.UserStatusActive {
		obslog.Init()
		obslog.L.WarnContext(ctx, obslog.MsgVerifyOTPRejected,
			append(authFlowAttrs(ctx, obslog.FlowVerifyOTP, obslog.StepAlreadyActive),
				"username", req.Username,
				"user_id", user.ID,
			)...,
		)
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "User already active"})
	}

	if user.OTPExpiresAt != nil && time.Now().After(*user.OTPExpiresAt) {
		obslog.Init()
		obslog.L.WarnContext(ctx, obslog.MsgVerifyOTPRejected,
			append(authFlowAttrs(ctx, obslog.FlowVerifyOTP, obslog.StepOTPExpired),
				"username", req.Username,
				"user_id", user.ID,
				"expires_at", user.OTPExpiresAt.Format(time.RFC3339),
			)...,
		)
		return c.JSON(http.StatusGone, map[string]string{"error": "OTP expired"})
	}

	if user.OTPCode != req.OTPCode {
		obslog.Init()
		obslog.L.WarnContext(ctx, obslog.MsgVerifyOTPRejected,
			append(authFlowAttrs(ctx, obslog.FlowVerifyOTP, obslog.StepOTPMismatch),
				"username", req.Username,
				"user_id", user.ID,
				"otp_len_provided", len(req.OTPCode),
			)...,
		)
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Invalid OTP code"})
	}

	// Activate user
	user.Status = models.UserStatusActive
	user.OTPCode = "" // Clear OTP after success
	user.OTPExpiresAt = nil
	if err := h.userRepo.Update(ctx, user); err != nil {
		obslog.Init()
		obslog.L.ErrorContext(ctx, obslog.MsgVerifyOTPFailed,
			append(authFlowAttrs(ctx, obslog.FlowVerifyOTP, obslog.StepDBUpdateUser),
				"username", req.Username,
				"user_id", user.ID,
				"err", err.Error(),
			)...,
		)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update user status"})
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "Account verified successfully"})
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// Login checks username/password for an active account and sends a one-time login code via SMS.
func (h *AuthHandler) Login(c echo.Context) error {
	var req LoginRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request payload"})
	}
	if req.Username == "" || req.Password == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Username and password are required"})
	}

	ctx := c.Request().Context()
	if h.limiter != nil {
		hour := time.Now().Unix() / 3600
		ipKey := fmt.Sprintf("login:1h:ip:%s:%d", httputil.ClientIP(c), hour)
		if err := h.limiter.Allow(ctx, ipKey, loginAttemptsPerIPPerHour, time.Hour); err != nil {
			return respondTooManyRequests(c, err)
		}
	}

	user, err := h.userRepo.GetByUsername(ctx, req.Username)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Database error"})
	}
	if user == nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Invalid username or password"})
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Invalid username or password"})
	}
	if user.Status != models.UserStatusActive {
		return c.JSON(http.StatusForbidden, map[string]string{"error": "Complete registration and verify your phone before signing in"})
	}

	rand.Seed(time.Now().UnixNano())
	otpCode := fmt.Sprintf("%06d", rand.Intn(1000000))
	otpTTL := getOTPTTL()
	expiresAt := time.Now().Add(otpTTL)
	user.OTPCode = otpCode
	user.OTPExpiresAt = &expiresAt
	if err := h.userRepo.Update(ctx, user); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to prepare login code"})
	}

	smsContent := fmt.Sprintf("Your login OTP is: %s", otpCode)
	smsCtx := sms.WithSendSource(sms.WithTestSMSMode(ctx, testSMSModeFromRequest(c)), sms.SendSourceAuth)
	_, err = h.smsService.SendSMS(smsCtx, user.Country, user.PhoneNumber, smsContent)
	if err != nil {
		user.OTPCode = ""
		user.OTPExpiresAt = nil
		_ = h.userRepo.Update(ctx, user)
		var lim *ratelimit.LimitedError
		if errors.As(err, &lim) {
			return respondTooManyRequests(c, err)
		}
		if errors.Is(err, phone.ErrInvalidCountry) || errors.Is(err, phone.ErrInvalidNumber) {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": fmt.Sprintf("Failed to send SMS: %v", err)})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"message":  "We sent a login code to your phone. Enter it to finish signing in.",
		"username": user.Username,
	})
}

type LoginVerifyRequest struct {
	Username string `json:"username"`
	OTPCode  string `json:"otp_code"`
}

// LoginVerify validates the SMS login code for an active user (separate from registration /api/verify).
func (h *AuthHandler) LoginVerify(c echo.Context) error {
	var req LoginVerifyRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request payload"})
	}
	if req.Username == "" || req.OTPCode == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Username and OTP code are required"})
	}

	ctx := c.Request().Context()
	if h.limiter != nil {
		hour := time.Now().Unix() / 3600
		ipKey := fmt.Sprintf("login_verify:1h:ip:%s:%d", httputil.ClientIP(c), hour)
		if err := h.limiter.Allow(ctx, ipKey, loginVerifyPerIPPerHour, time.Hour); err != nil {
			return respondTooManyRequests(c, err)
		}
	}

	user, err := h.userRepo.GetByUsername(ctx, req.Username)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Database error"})
	}
	if user == nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Invalid username or login code"})
	}
	if user.Status != models.UserStatusActive {
		return c.JSON(http.StatusForbidden, map[string]string{"error": "Account is not active"})
	}
	if user.OTPCode == "" || user.OTPExpiresAt == nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "No active login code. Sign in with your password again."})
	}
	if time.Now().After(*user.OTPExpiresAt) {
		return c.JSON(http.StatusGone, map[string]string{"error": "Login code expired"})
	}
	if user.OTPCode != req.OTPCode {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Invalid username or login code"})
	}

	user.OTPCode = ""
	user.OTPExpiresAt = nil
	if err := h.userRepo.Update(ctx, user); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to complete sign-in"})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"message": "Signed in successfully",
		"user": map[string]interface{}{
			"id":           user.ID,
			"username":     user.Username,
			"phone_number": user.PhoneNumber,
			"country":      user.Country,
			"status":       user.Status,
		},
	})
}
