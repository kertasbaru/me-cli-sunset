package handler

import (
	"net/http"

	"github.com/kertasbaru/me-cli-sunset/internal/client"
	"github.com/kertasbaru/me-cli-sunset/internal/model"
	"github.com/kertasbaru/me-cli-sunset/internal/service"
)

// FamilyHandler handles family plan HTTP requests.
type FamilyHandler struct {
	authService *service.AuthService
	client      *client.Client
}

// NewFamilyHandler creates a new FamilyHandler.
func NewFamilyHandler(authService *service.AuthService, cl *client.Client) *FamilyHandler {
	return &FamilyHandler{
		authService: authService,
		client:      cl,
	}
}

// GetFamilyPlanData handles GET /api/v1/family/plan
func (h *FamilyHandler) GetFamilyPlanData(w http.ResponseWriter, r *http.Request) {
	tokens := h.authService.GetActiveTokens()
	if tokens == nil {
		writeError(w, http.StatusUnauthorized, "no active session")
		return
	}

	data, err := h.client.GetFamilyPlanData(tokens.IDToken)
	if err != nil {
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}

	writeSuccess(w, data)
}

// ValidateMSISDN handles POST /api/v1/family/validate
func (h *FamilyHandler) ValidateMSISDN(w http.ResponseWriter, r *http.Request) {
	tokens := h.authService.GetActiveTokens()
	if tokens == nil {
		writeError(w, http.StatusUnauthorized, "no active session")
		return
	}

	var req struct {
		MSISDN string `json:"msisdn"`
	}
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	data, err := h.client.ValidateFamplanMSISDN(tokens.IDToken, req.MSISDN)
	if err != nil {
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}

	writeSuccess(w, data)
}

// ChangeMember handles POST /api/v1/family/members/change
func (h *FamilyHandler) ChangeMember(w http.ResponseWriter, r *http.Request) {
	tokens := h.authService.GetActiveTokens()
	if tokens == nil {
		writeError(w, http.StatusUnauthorized, "no active session")
		return
	}

	var req struct {
		ParentAlias    string `json:"parent_alias"`
		Alias          string `json:"alias"`
		SlotID         int    `json:"slot_id"`
		FamilyMemberID string `json:"family_member_id"`
		NewMSISDN      string `json:"new_msisdn"`
	}
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	data, err := h.client.ChangeFamplanMember(tokens.IDToken, req.ParentAlias, req.Alias, req.SlotID, req.FamilyMemberID, req.NewMSISDN)
	if err != nil {
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}

	writeSuccess(w, data)
}

// RemoveMember handles POST /api/v1/family/members/remove
func (h *FamilyHandler) RemoveMember(w http.ResponseWriter, r *http.Request) {
	tokens := h.authService.GetActiveTokens()
	if tokens == nil {
		writeError(w, http.StatusUnauthorized, "no active session")
		return
	}

	var req struct {
		FamilyMemberID string `json:"family_member_id"`
	}
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	data, err := h.client.RemoveFamplanMember(tokens.IDToken, req.FamilyMemberID)
	if err != nil {
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}

	writeSuccess(w, data)
}

// SetQuotaLimit handles POST /api/v1/family/quota
func (h *FamilyHandler) SetQuotaLimit(w http.ResponseWriter, r *http.Request) {
	tokens := h.authService.GetActiveTokens()
	if tokens == nil {
		writeError(w, http.StatusUnauthorized, "no active session")
		return
	}

	var req model.FamplanQuotaRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	data, err := h.client.SetFamplanQuotaLimit(tokens.IDToken, req.OriginalAllocation, req.NewAllocation, req.FamilyMemberID)
	if err != nil {
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}

	writeSuccess(w, data)
}
