package handler

import (
	"net/http"
	"strconv"

	"github.com/kertasbaru/me-cli-sunset/internal/database"
	"github.com/kertasbaru/me-cli-sunset/internal/model"
	"github.com/kertasbaru/me-cli-sunset/internal/service"
)

// BookmarkHandler handles bookmark HTTP requests.
type BookmarkHandler struct {
	authService *service.AuthService
	db          *database.DB
}

// NewBookmarkHandler creates a new BookmarkHandler.
func NewBookmarkHandler(authService *service.AuthService, db *database.DB) *BookmarkHandler {
	return &BookmarkHandler{
		authService: authService,
		db:          db,
	}
}

// ListBookmarks handles GET /api/v1/bookmarks
func (h *BookmarkHandler) ListBookmarks(w http.ResponseWriter, r *http.Request) {
	user := h.authService.GetActiveUser()
	if user == nil {
		writeError(w, http.StatusUnauthorized, "no active session")
		return
	}

	account, err := h.db.GetAccountByNumber(user.Number)
	if err != nil || account == nil {
		writeError(w, http.StatusInternalServerError, "failed to get account")
		return
	}

	bookmarks, err := h.db.GetBookmarks(account.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if bookmarks == nil {
		bookmarks = []model.Bookmark{}
	}

	writeSuccess(w, bookmarks)
}

// AddBookmark handles POST /api/v1/bookmarks
func (h *BookmarkHandler) AddBookmark(w http.ResponseWriter, r *http.Request) {
	user := h.authService.GetActiveUser()
	if user == nil {
		writeError(w, http.StatusUnauthorized, "no active session")
		return
	}

	account, err := h.db.GetAccountByNumber(user.Number)
	if err != nil || account == nil {
		writeError(w, http.StatusInternalServerError, "failed to get account")
		return
	}

	var req model.BookmarkRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.FamilyCode == "" {
		writeError(w, http.StatusBadRequest, "family_code is required")
		return
	}

	if err := h.db.AddBookmark(account.ID, req.FamilyCode, req.FamilyName, req.IsEnterprise, req.VariantName, req.OptionName, req.Order); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeSuccess(w, map[string]string{
		"message": "bookmark added",
	})
}

// DeleteBookmark handles DELETE /api/v1/bookmarks/{id}
func (h *BookmarkHandler) DeleteBookmark(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid bookmark id")
		return
	}

	if err := h.db.DeleteBookmark(id); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeSuccess(w, map[string]string{
		"message": "bookmark deleted",
	})
}
