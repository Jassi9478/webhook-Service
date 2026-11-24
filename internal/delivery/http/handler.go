package http

import (
	"fmt"
	"io"
	"net/http"
	"time"
	"webhook-service/internal/usecase"
)

type IngestHandler struct {
	useCase *usecase.IngestEventUseCase
}

func NewIngestHandler(uc *usecase.IngestEventUseCase) *IngestHandler {
	return &IngestHandler{useCase: uc}
}

func (h *IngestHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	tenantID := r.Header.Get("X-Tenant-ID")
	eventType := r.Header.Get("X-Event-Type")
	idempotencyKey := r.Header.Get("X-Idempotency-Key")

	body, err := io.ReadAll(r.Body)
	if err != nil || len(body) == 0 {
		http.Error(w, "Invalid payload body", http.StatusBadRequest)
		return
	}

	eventID := fmt.Sprintf("evt_%d", time.Now().UnixNano())
	event, err := h.useCase.Execute(r.Context(), eventID, tenantID, eventType, idempotencyKey, string(body))
	if err != nil {
		http.Error(w, err.Error(), http.StatusConflict)
		return
	}

	w.WriteHeader(http.StatusAccepted)
	w.Write([]byte(fmt.Sprintf(`{"status":"accepted","event_id":"%s"}`, event.ID)))
}
