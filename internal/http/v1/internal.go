package v1

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/go-chi/chi/v5"
	"net/http"
)

func (h *Handler) initInternalRoutes(r chi.Router) {
	r.Route("/internal", func(r chi.Router) {
		r.Get("/stats", h.GetStats())
	})
}

// GetUserLinks returns to the user all links ever saved by him.
func (h *Handler) GetStats() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ipStr := r.Header.Get("X-Real-IP")

		//TODO move it in middleware
		trusted, err := h.services.Security.IsIpAddrTrusted(r.Context(), ipStr)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if !trusted {
			http.Error(w, "the ip address is not on a trusted network", http.StatusForbidden)
			return
		}

		stat, err := h.services.Internal.GetStatistic(r.Context())
		if err != nil {
			http.Error(w, err.Error(), http.StatusNoContent)
			return
		}

		buf := bytes.NewBuffer([]byte{})
		encoder := json.NewEncoder(buf)
		encoder.SetEscapeHTML(false)
		if err = encoder.Encode(stat); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("content-type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, buf)
	}
}
