package handler

import (
	"net/http"

	"github.com/kertasbaru/me-cli-sunset/internal/client"
	"github.com/kertasbaru/me-cli-sunset/internal/model"
	"github.com/kertasbaru/me-cli-sunset/internal/service"
)

// PackageHandler handles package-related HTTP requests.
type PackageHandler struct {
	authService *service.AuthService
	client      *client.Client
}

// NewPackageHandler creates a new PackageHandler.
func NewPackageHandler(authService *service.AuthService, cl *client.Client) *PackageHandler {
	return &PackageHandler{
		authService: authService,
		client:      cl,
	}
}

// GetBalance handles GET /api/v1/packages/balance
func (h *PackageHandler) GetBalance(w http.ResponseWriter, r *http.Request) {
	tokens := h.authService.GetActiveTokens()
	if tokens == nil {
		writeError(w, http.StatusUnauthorized, "no active session")
		return
	}

	balance, err := h.client.GetBalance(tokens.IDToken)
	if err != nil {
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}

	writeSuccess(w, balance)
}

// GetFamily handles POST /api/v1/packages/family
func (h *PackageHandler) GetFamily(w http.ResponseWriter, r *http.Request) {
	tokens := h.authService.GetActiveTokens()
	if tokens == nil {
		writeError(w, http.StatusUnauthorized, "no active session")
		return
	}

	var req model.FamilyRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.FamilyCode == "" {
		writeError(w, http.StatusBadRequest, "family_code is required")
		return
	}

	var migrationType *string
	if req.MigrationType != "" {
		migrationType = &req.MigrationType
	}

	data, err := h.client.GetFamily(tokens.IDToken, req.FamilyCode, req.IsEnterprise, migrationType)
	if err != nil {
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}

	if data == nil {
		writeError(w, http.StatusNotFound, "family not found")
		return
	}

	writeSuccess(w, data)
}

// GetFamilies handles GET /api/v1/packages/families
func (h *PackageHandler) GetFamilies(w http.ResponseWriter, r *http.Request) {
	tokens := h.authService.GetActiveTokens()
	if tokens == nil {
		writeError(w, http.StatusUnauthorized, "no active session")
		return
	}

	categoryCode := r.URL.Query().Get("category_code")
	if categoryCode == "" {
		writeError(w, http.StatusBadRequest, "category_code query parameter is required")
		return
	}

	data, err := h.client.GetFamilies(tokens.IDToken, categoryCode)
	if err != nil {
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}

	writeSuccess(w, data)
}

// GetPackageDetail handles POST /api/v1/packages/detail
func (h *PackageHandler) GetPackageDetail(w http.ResponseWriter, r *http.Request) {
	tokens := h.authService.GetActiveTokens()
	if tokens == nil {
		writeError(w, http.StatusUnauthorized, "no active session")
		return
	}

	var req model.PackageRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.PackageOptionCode == "" {
		writeError(w, http.StatusBadRequest, "package_option_code is required")
		return
	}

	data, err := h.client.GetPackage(tokens.IDToken, req.PackageOptionCode, req.PackageFamilyCode, req.PackageVariantCode)
	if err != nil {
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}

	if data == nil {
		writeError(w, http.StatusNotFound, "package not found")
		return
	}

	writeSuccess(w, data)
}

// GetAddons handles GET /api/v1/packages/addons
func (h *PackageHandler) GetAddons(w http.ResponseWriter, r *http.Request) {
	tokens := h.authService.GetActiveTokens()
	if tokens == nil {
		writeError(w, http.StatusUnauthorized, "no active session")
		return
	}

	optionCode := r.URL.Query().Get("option_code")
	if optionCode == "" {
		writeError(w, http.StatusBadRequest, "option_code query parameter is required")
		return
	}

	data, err := h.client.GetAddons(tokens.IDToken, optionCode)
	if err != nil {
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}

	writeSuccess(w, data)
}

// GetTieringInfo handles GET /api/v1/packages/tiering
func (h *PackageHandler) GetTieringInfo(w http.ResponseWriter, r *http.Request) {
	tokens := h.authService.GetActiveTokens()
	if tokens == nil {
		writeError(w, http.StatusUnauthorized, "no active session")
		return
	}

	data, err := h.client.GetTieringInfo(tokens.IDToken)
	if err != nil {
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}

	writeSuccess(w, data)
}

// GetQuotaDetails handles GET /api/v1/packages/quotas
func (h *PackageHandler) GetQuotaDetails(w http.ResponseWriter, r *http.Request) {
	tokens := h.authService.GetActiveTokens()
	if tokens == nil {
		writeError(w, http.StatusUnauthorized, "no active session")
		return
	}

	data, err := h.client.GetQuotaDetails(tokens.IDToken)
	if err != nil {
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}

	writeSuccess(w, data)
}

// GetPaymentMethods handles POST /api/v1/packages/payment-methods
func (h *PackageHandler) GetPaymentMethods(w http.ResponseWriter, r *http.Request) {
	tokens := h.authService.GetActiveTokens()
	if tokens == nil {
		writeError(w, http.StatusUnauthorized, "no active session")
		return
	}

	var req struct {
		TokenConfirmation string `json:"token_confirmation"`
		PaymentTarget     string `json:"payment_target"`
	}
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	data, err := h.client.GetPaymentMethods(tokens.IDToken, req.TokenConfirmation, req.PaymentTarget)
	if err != nil {
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}

	writeSuccess(w, data)
}

// Unsubscribe handles POST /api/v1/packages/unsubscribe
func (h *PackageHandler) Unsubscribe(w http.ResponseWriter, r *http.Request) {
	tokens := h.authService.GetActiveTokens()
	if tokens == nil {
		writeError(w, http.StatusUnauthorized, "no active session")
		return
	}

	var req struct {
		QuotaCode               string `json:"quota_code"`
		ProductDomain           string `json:"product_domain"`
		ProductSubscriptionType string `json:"product_subscription_type"`
	}
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	success, err := h.client.Unsubscribe(tokens.IDToken, req.QuotaCode, req.ProductDomain, req.ProductSubscriptionType)
	if err != nil {
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}

	writeSuccess(w, map[string]bool{
		"success": success,
	})
}
