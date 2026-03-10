package handler

import (
	"net/http"
	"strconv"

	"github.com/kertasbaru/me-cli-sunset/internal/client"
	"github.com/kertasbaru/me-cli-sunset/internal/model"
	"github.com/kertasbaru/me-cli-sunset/internal/service"
)

// AuthHandler handles authentication-related HTTP requests.
type AuthHandler struct {
	authService *service.AuthService
	client      *client.Client
}

// NewAuthHandler creates a new AuthHandler.
func NewAuthHandler(authService *service.AuthService, cl *client.Client) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		client:      cl,
	}
}

// RequestOTP handles POST /api/v1/auth/otp/request
func (h *AuthHandler) RequestOTP(w http.ResponseWriter, r *http.Request) {
	var req model.OTPRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Contact == "" {
		writeError(w, http.StatusBadRequest, "contact is required")
		return
	}

	subscriberID, err := h.client.RequestOTP(req.Contact)
	if err != nil {
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}

	writeSuccess(w, map[string]string{
		"subscriber_id": subscriberID,
	})
}

// SubmitOTP handles POST /api/v1/auth/otp/submit
func (h *AuthHandler) SubmitOTP(w http.ResponseWriter, r *http.Request) {
	var req model.OTPSubmitRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Contact == "" || req.Code == "" {
		writeError(w, http.StatusBadRequest, "contact and code are required")
		return
	}

	contactType := req.ContactType
	if contactType == "" {
		contactType = "SMS"
	}

	tokens, err := h.client.SubmitOTP(contactType, req.Contact, req.Code)
	if err != nil {
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}

	refreshToken, _ := tokens["refresh_token"].(string)
	number, err := strconv.ParseInt(req.Contact, 10, 64)
	if err == nil && refreshToken != "" {
		if err := h.authService.AddAccount(number, refreshToken); err != nil {
			writeError(w, http.StatusInternalServerError, "login succeeded but failed to save account: "+err.Error())
			return
		}
	}

	writeSuccess(w, tokens)
}

// ListAccounts handles GET /api/v1/auth/accounts
func (h *AuthHandler) ListAccounts(w http.ResponseWriter, r *http.Request) {
	accounts, err := h.authService.GetAllAccounts()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeSuccess(w, accounts)
}

// SwitchAccount handles POST /api/v1/auth/accounts/switch
func (h *AuthHandler) SwitchAccount(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Number int64 `json:"number"`
	}
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := h.authService.SetActiveAccount(req.Number); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeSuccess(w, map[string]interface{}{
		"message": "account switched successfully",
		"number":  req.Number,
	})
}

// DeleteAccount handles DELETE /api/v1/auth/accounts/{number}
func (h *AuthHandler) DeleteAccount(w http.ResponseWriter, r *http.Request) {
	numberStr := r.PathValue("number")
	number, err := strconv.ParseInt(numberStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid number")
		return
	}

	if err := h.authService.RemoveAccount(number); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeSuccess(w, map[string]string{
		"message": "account deleted successfully",
	})
}

// GetActiveUser handles GET /api/v1/auth/active
func (h *AuthHandler) GetActiveUser(w http.ResponseWriter, r *http.Request) {
	user := h.authService.GetActiveUser()
	if user == nil {
		writeError(w, http.StatusNotFound, "no active user")
		return
	}

	writeSuccess(w, map[string]interface{}{
		"number":            user.Number,
		"subscriber_id":     user.SubscriberID,
		"subscription_type": user.SubscriptionType,
	})
}

// RenewToken handles POST /api/v1/auth/renew
func (h *AuthHandler) RenewToken(w http.ResponseWriter, r *http.Request) {
	user := h.authService.GetActiveUser()
	if user == nil {
		writeError(w, http.StatusUnauthorized, "no active user")
		return
	}

	writeSuccess(w, map[string]string{
		"message": "token renewed successfully",
	})
}
