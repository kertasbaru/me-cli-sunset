package handler

import (
	"net/http"

	"github.com/kertasbaru/me-cli-sunset/internal/client"
	"github.com/kertasbaru/me-cli-sunset/internal/service"
)

// NotificationHandler handles notification HTTP requests.
type NotificationHandler struct {
	authService *service.AuthService
	client      *client.Client
}

// NewNotificationHandler creates a new NotificationHandler.
func NewNotificationHandler(authService *service.AuthService, cl *client.Client) *NotificationHandler {
	return &NotificationHandler{
		authService: authService,
		client:      cl,
	}
}

// GetNotifications handles GET /api/v1/notifications
func (h *NotificationHandler) GetNotifications(w http.ResponseWriter, r *http.Request) {
	tokens := h.authService.GetActiveTokens()
	if tokens == nil {
		writeError(w, http.StatusUnauthorized, "no active session")
		return
	}

	data, err := h.client.GetNotifications(tokens.IDToken)
	if err != nil {
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}

	writeSuccess(w, data)
}

// GetNotificationDetail handles GET /api/v1/notifications/{notification_id}
func (h *NotificationHandler) GetNotificationDetail(w http.ResponseWriter, r *http.Request) {
	tokens := h.authService.GetActiveTokens()
	if tokens == nil {
		writeError(w, http.StatusUnauthorized, "no active session")
		return
	}

	notificationID := r.PathValue("notification_id")
	if notificationID == "" {
		writeError(w, http.StatusBadRequest, "notification_id is required")
		return
	}

	data, err := h.client.GetNotificationDetail(tokens.IDToken, notificationID)
	if err != nil {
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}

	writeSuccess(w, data)
}

// GetTransactionHistory handles GET /api/v1/transactions
func (h *NotificationHandler) GetTransactionHistory(w http.ResponseWriter, r *http.Request) {
	tokens := h.authService.GetActiveTokens()
	if tokens == nil {
		writeError(w, http.StatusUnauthorized, "no active session")
		return
	}

	data, err := h.client.GetTransactionHistory(tokens.IDToken)
	if err != nil {
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}

	writeSuccess(w, data)
}

// GetDashboard handles GET /api/v1/dashboard
func (h *NotificationHandler) GetDashboard(w http.ResponseWriter, r *http.Request) {
	user := h.authService.GetActiveUser()
	if user == nil {
		writeError(w, http.StatusUnauthorized, "no active session")
		return
	}

	data, err := h.client.DashboardSegments(user.Tokens.AccessToken, user.Tokens.IDToken)
	if err != nil {
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}

	writeSuccess(w, data)
}
