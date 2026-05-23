package baseapi

import (
	"log/slog"
	"net/http"
	"todo-runner-cli/internal/httpresponse"
)

type JSONWriter interface {
	WriteJSON(w http.ResponseWriter, statusCode int, payload any)
}

type BaseHandler struct {
	logger   *slog.Logger
	httpUtil JSONWriter
}

func NewBaseHandler(logger *slog.Logger, httpUtil JSONWriter) *BaseHandler {
	return &BaseHandler{
		logger:   logger,
		httpUtil: httpUtil,
	}
}

func (t *BaseHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/health", t.healthHandler)
}

func (t *BaseHandler) healthHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		t.httpUtil.WriteJSON(w, http.StatusMethodNotAllowed, httpresponse.ErrorResponse{Error: "method: " + r.Method + " is not allowed"})
		return
	}
	t.httpUtil.WriteJSON(w, http.StatusOK, httpresponse.SuccessResponse{Data: httpresponse.StatusResponse{Status: "ok"}})
}
