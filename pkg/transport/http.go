package transport

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/mohammadhptp/pulse/internal/storage"
	"github.com/mohammadhptp/pulse/pkg/logger"
	"github.com/mohammadhptp/pulse/pkg/models"
	"go.uber.org/zap"
)

type HTTPTransport struct {
	server   *http.Server
	handler  EventHandler
	port     int
	endpoint string
	mu       sync.RWMutex
}

func NewHTTPTransport(port int, endpoint string) *HTTPTransport {
	return &HTTPTransport{
		port:     port,
		endpoint: endpoint,
	}
}

func (h *HTTPTransport) SetEventHandler(handler EventHandler) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.handler = handler
}

func (h *HTTPTransport) Start(ctx context.Context) error {
	mux := http.NewServeMux()
	mux.HandleFunc(h.endpoint, h.handleEvents)
	mux.HandleFunc(h.endpoint, h.handleFilterEvents)

	addr := fmt.Sprintf(":%d", h.port)
	h.server = &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	logger.Info("Starting HTTP transport", zap.String("address", addr), zap.String("endpoint", h.endpoint))

	go func() {
		if err := h.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("HTTP server error", zap.Error(err))
		}
	}()

	go func() {
		<-ctx.Done()
		h.Stop()
	}()

	return nil
}

func (h *HTTPTransport) Stop() error {
	if h.server == nil {
		return nil
	}

	logger.Info("Stopping HTTP transport")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return h.server.Shutdown(ctx)
}

func (h *HTTPTransport) Close() error {
	return h.Stop()
}

func (h *HTTPTransport) handleEvents(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	h.mu.RLock()
	handler := h.handler
	h.mu.RUnlock()

	if handler == nil {
		http.Error(w, "Event handler not configured", http.StatusInternalServerError)
		return
	}

	var event models.Event
	event.RequestID = uuid.New().String()
	if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
		logger.Warn("Failed to parse event", zap.Error(err))
		http.Error(w, "Bad request: "+err.Error(), http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	if err := handler(event); err != nil {
		logger.Error("Failed to process event", zap.Error(err))
		http.Error(w, "Failed to process event", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusAccepted)
	w.Write([]byte(`{"status":"accepted"}`))
}

func (h *HTTPTransport) handleFilterEvents(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	opts := models.QueryOptions{}

	query := r.URL.Query()
	opts.Service = query.Get("service")
	opts.Level = query.Get("level")
	opts.Host = query.Get("host")
	opts.RequestID = query.Get("request_id")
	opts.SearchQuery = query.Get("search")
	opts.SortOrder = query.Get("sort_order")

	if v := query.Get("limit"); v != "" {
		var limit int
		if _, err := fmt.Sscanf(v, "%d", &limit); err == nil && limit > 0 {
			opts.Limit = limit
		}
	}

	if v := query.Get("offset"); v != "" {
		var offset int
		if _, err := fmt.Sscanf(v, "%d", &offset); err == nil && offset >= 0 {
			opts.Offset = offset
		}
	}

	if v := query.Get("start_time"); v != "" {
		var startTime uint64
		if _, err := fmt.Sscanf(v, "%d", &startTime); err == nil {
			opts.StartTime = startTime
		}
	}

	if v := query.Get("end_time"); v != "" {
		var endTime uint64
		if _, err := fmt.Sscanf(v, "%d", &endTime); err == nil {
			opts.EndTime = endTime
		}
	}

	conn, err := storage.Connect(context.Background())
	if err != nil {
		logger.Error("Failed to connect to ClickHouse", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	events, err := storage.QueryEvents(r.Context(), conn, opts)
	if err != nil {
		logger.Error("Failed to query events", zap.Error(err))
		http.Error(w, "Failed to query events", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(events); err != nil {
		logger.Error("Failed to encode response", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}
