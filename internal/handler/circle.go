package handler

import (
	"net/http"

	"github.com/kertasbaru/me-cli-sunset/internal/client"
	"github.com/kertasbaru/me-cli-sunset/internal/model"
	"github.com/kertasbaru/me-cli-sunset/internal/service"
)

// CircleHandler handles family circle HTTP requests.
type CircleHandler struct {
	authService *service.AuthService
	client      *client.Client
}

// NewCircleHandler creates a new CircleHandler.
func NewCircleHandler(authService *service.AuthService, cl *client.Client) *CircleHandler {
	return &CircleHandler{
		authService: authService,
		client:      cl,
	}
}

// GetGroupData handles GET /api/v1/circle/groups
func (h *CircleHandler) GetGroupData(w http.ResponseWriter, r *http.Request) {
	tokens := h.authService.GetActiveTokens()
	if tokens == nil {
		writeError(w, http.StatusUnauthorized, "no active session")
		return
	}

	data, err := h.client.GetGroupData(tokens.IDToken)
	if err != nil {
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}

	writeSuccess(w, data)
}

// GetGroupMembers handles GET /api/v1/circle/groups/{group_id}/members
func (h *CircleHandler) GetGroupMembers(w http.ResponseWriter, r *http.Request) {
	tokens := h.authService.GetActiveTokens()
	if tokens == nil {
		writeError(w, http.StatusUnauthorized, "no active session")
		return
	}

	groupID := r.PathValue("group_id")
	if groupID == "" {
		writeError(w, http.StatusBadRequest, "group_id is required")
		return
	}

	data, err := h.client.GetGroupMembers(tokens.IDToken, groupID)
	if err != nil {
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}

	writeSuccess(w, data)
}

// CreateCircle handles POST /api/v1/circle/groups
func (h *CircleHandler) CreateCircle(w http.ResponseWriter, r *http.Request) {
	tokens := h.authService.GetActiveTokens()
	if tokens == nil {
		writeError(w, http.StatusUnauthorized, "no active session")
		return
	}

	var req model.CircleCreateRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	data, err := h.client.CreateCircle(tokens.AccessToken, tokens.IDToken, req.ParentName, req.GroupName, req.MemberMSISDN, req.MemberName)
	if err != nil {
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}

	writeSuccess(w, data)
}

// ValidateMember handles POST /api/v1/circle/members/validate
func (h *CircleHandler) ValidateMember(w http.ResponseWriter, r *http.Request) {
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

	data, err := h.client.ValidateCircleMember(tokens.IDToken, req.MSISDN)
	if err != nil {
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}

	writeSuccess(w, data)
}

// InviteMember handles POST /api/v1/circle/members/invite
func (h *CircleHandler) InviteMember(w http.ResponseWriter, r *http.Request) {
	tokens := h.authService.GetActiveTokens()
	if tokens == nil {
		writeError(w, http.StatusUnauthorized, "no active session")
		return
	}

	var req model.CircleMemberRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	data, err := h.client.InviteCircleMember(tokens.AccessToken, tokens.IDToken, req.MSISDN, req.Name, req.GroupID, req.MemberIDParent)
	if err != nil {
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}

	writeSuccess(w, data)
}

// RemoveMember handles POST /api/v1/circle/members/remove
func (h *CircleHandler) RemoveMember(w http.ResponseWriter, r *http.Request) {
	tokens := h.authService.GetActiveTokens()
	if tokens == nil {
		writeError(w, http.StatusUnauthorized, "no active session")
		return
	}

	var req struct {
		MemberID       string `json:"member_id"`
		GroupID        string `json:"group_id"`
		MemberIDParent string `json:"member_id_parent"`
		IsLastMember   bool   `json:"is_last_member"`
	}
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	data, err := h.client.RemoveCircleMember(tokens.IDToken, req.MemberID, req.GroupID, req.MemberIDParent, req.IsLastMember)
	if err != nil {
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}

	writeSuccess(w, data)
}

// AcceptInvitation handles POST /api/v1/circle/invitations/accept
func (h *CircleHandler) AcceptInvitation(w http.ResponseWriter, r *http.Request) {
	tokens := h.authService.GetActiveTokens()
	if tokens == nil {
		writeError(w, http.StatusUnauthorized, "no active session")
		return
	}

	var req struct {
		GroupID  string `json:"group_id"`
		MemberID string `json:"member_id"`
	}
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	data, err := h.client.AcceptCircleInvitation(tokens.AccessToken, tokens.IDToken, req.GroupID, req.MemberID)
	if err != nil {
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}

	writeSuccess(w, data)
}

// GetSpendingTracker handles POST /api/v1/circle/spending-tracker
func (h *CircleHandler) GetSpendingTracker(w http.ResponseWriter, r *http.Request) {
	tokens := h.authService.GetActiveTokens()
	if tokens == nil {
		writeError(w, http.StatusUnauthorized, "no active session")
		return
	}

	var req struct {
		ParentSubsID string `json:"parent_subs_id"`
		FamilyID     string `json:"family_id"`
	}
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	data, err := h.client.SpendingTracker(tokens.IDToken, req.ParentSubsID, req.FamilyID)
	if err != nil {
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}

	writeSuccess(w, data)
}

// GetBonusData handles POST /api/v1/circle/bonus
func (h *CircleHandler) GetBonusData(w http.ResponseWriter, r *http.Request) {
	tokens := h.authService.GetActiveTokens()
	if tokens == nil {
		writeError(w, http.StatusUnauthorized, "no active session")
		return
	}

	var req struct {
		ParentSubsID string `json:"parent_subs_id"`
		FamilyID     string `json:"family_id"`
	}
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	data, err := h.client.GetBonusData(tokens.IDToken, req.ParentSubsID, req.FamilyID)
	if err != nil {
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}

	writeSuccess(w, data)
}
