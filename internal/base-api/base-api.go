package base_api

import (
	"encoding/json"
	"log/slog"
	"net/http"
)

type ErrorResponse struct {
	Error string `json:"error"`
}

type SuccessResponse struct {
	Status string `json:"status"`
}

type BaseHandler struct {
	logger *slog.Logger
}

func NewBaseHandler(logger *slog.Logger) *BaseHandler {
	return &BaseHandler{
		logger: logger,
	}
}

func (t *BaseHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/health", t.healthHandler)
}

func (t *BaseHandler) writeJSON(w http.ResponseWriter, statusCode int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(payload); err != nil {
		t.logger.Error("failed to write json response", "error", err)
	}
}

func (t *BaseHandler) healthHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		t.writeJSON(w, http.StatusMethodNotAllowed, ErrorResponse{Error: "method: " + r.Method + " is not allowed"})
		return
	}
	t.writeJSON(w, http.StatusOK, SuccessResponse{Status: "ok"})
}
