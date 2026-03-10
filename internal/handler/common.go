// Package handler provides HTTP request handlers for the REST API.
package handler

import (
	"encoding/json"
	"net/http"

	"github.com/kertasbaru/me-cli-sunset/internal/model"
)

// writeJSON writes a JSON response with the given status code.
func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// writeSuccess writes a success response with data.
func writeSuccess(w http.ResponseWriter, data interface{}) {
	writeJSON(w, http.StatusOK, model.APIResponse{
		Status: "success",
		Data:   data,
	})
}

// writeError writes an error response.
func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, model.APIResponse{
		Status: "error",
		Error:  message,
	})
}

// decodeJSON decodes a JSON request body into the given target.
func decodeJSON(r *http.Request, target interface{}) error {
	return json.NewDecoder(r.Body).Decode(target)
}
