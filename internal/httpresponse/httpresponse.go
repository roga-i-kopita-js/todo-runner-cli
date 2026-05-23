package httpresponse

import (
	"encoding/json"
	"log/slog"
	"net/http"
)

type ErrorResponse struct {
	Error string `json:"error"`
}

type StatusResponse struct {
	Status string `json:"status"`
}

type SuccessResponse struct {
	Data any `json:"data"`
}

type JSONWriter struct {
	logger *slog.Logger
}

func NewJSONWriter(logger *slog.Logger) *JSONWriter {
	return &JSONWriter{
		logger: logger,
	}
}

func (t *JSONWriter) WriteJSON(w http.ResponseWriter, statusCode int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(payload); err != nil {
		t.logger.Error("failed to write json response", "error", err)
	}
}
