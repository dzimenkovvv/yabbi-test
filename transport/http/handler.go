package http

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"yabbi_test/internal/auction"
)

type auctionService interface {
	RunAuction(ctx context.Context, req auction.Request) auction.Response
}

type Handler struct {
	service auctionService
	logger  *slog.Logger
}

func NewHandler(service auctionService, logger *slog.Logger) *Handler {
	return &Handler{service: service, logger: logger}
}

func (h *Handler) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/auction", h.handleAuction)
	return h.recoverMiddleware(mux)
}

func (h *Handler) handleAuction(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "only POST is supported")
		return
	}

	defer r.Body.Close()

	var req auction.Request
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body: "+err.Error())
		return
	}

	if err := auction.Validate(req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	resp := h.service.RunAuction(r.Context(), req)
	writeJSON(w, http.StatusOK, resp)
}

func (h *Handler) recoverMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				h.logger.Error("panic recovered in http handler",
					"panic", rec,
					"path", r.URL.Path,
				)
				writeError(w, http.StatusInternalServerError, "internal server error")
			}
		}()
		next.ServeHTTP(w, r)
	})
}

type apiError struct {
	Error string `json:"error"`
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, apiError{Error: message})
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(body); err != nil {
		// На этом этапе заголовки и статус уже отправлены клиенту,
		// вернуть корректную ошибку невозможно — только залогировать.
		_ = errors.New("response already committed, cannot report encode error to client")
	}
}
