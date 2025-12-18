package api

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	
	"mysql-sync-service/internal/sync"
)

type Handler struct {
	syncManager *sync.Manager
}

func NewHandler(manager *sync.Manager) *Handler {
	return &Handler{
		syncManager: manager,
	}
}

func (h *Handler) Routes() chi.Router {
	r := chi.NewRouter()
	
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(CorsMiddleware)
	
	r.Get("/health", h.HealthCheck)
	
	r.Route("/api/v1", func(r chi.Router) {
		r.Use(AuthMiddleware) // Placeholder for auth
		
		r.Post("/sync/trigger", h.TriggerSync)
		r.Post("/sync/stop", h.StopSync)
		r.Get("/sync/status", h.GetSyncStatus)
		// Add other routes
	})
	
	return r
}

func (h *Handler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func (h *Handler) TriggerSync(w http.ResponseWriter, r *http.Request) {
	if err := h.syncManager.Start(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	json.NewEncoder(w).Encode(map[string]string{"status": "started"})
}

func (h *Handler) StopSync(w http.ResponseWriter, r *http.Request) {
	h.syncManager.Stop()
	json.NewEncoder(w).Encode(map[string]string{"status": "stopped"})
}

func (h *Handler) GetSyncStatus(w http.ResponseWriter, r *http.Request) {
	status := h.syncManager.GetStatus()
	json.NewEncoder(w).Encode(map[string]string{"status": status})
}

// Middleware placeholders
func CorsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Authorization, Content-Type, X-CSRF-Token")
		
		if r.Method == "OPTIONS" {
			return
		}
		
		next.ServeHTTP(w, r)
	})
}

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// TODO: Implement actual token check
		next.ServeHTTP(w, r)
	})
}
