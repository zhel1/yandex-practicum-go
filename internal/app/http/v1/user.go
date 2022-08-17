package v1

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/zhel1/yandex-practicum-go/internal/app/http/middleware"
	"net/http"
)

func (h *Handler) initUserRoutes(r chi.Router) {
	r.Route("/user", func(r chi.Router) {
		r.Get("/urls", h.GetUserLinks())
		r.Delete("/urls", h.DeleteUserLinksBatch())
	})
}

//GetUserLinks returns to the user all links ever saved by him
func (h *Handler) GetUserLinks() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, err := middleware.TakeUserID(r.Context())
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		modelURLs, err := h.services.Users.GetURLsByUserID(r.Context(), userID)
		if err != nil || len(modelURLs) == 0 {
			http.Error(w, err.Error(), http.StatusNoContent)
			return
		}

		buf := bytes.NewBuffer([]byte{})
		encoder := json.NewEncoder(buf)
		encoder.SetEscapeHTML(false)
		if err = encoder.Encode(modelURLs); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("content-type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, buf)
	}
}

//DeleteUserLinksBatch accepts a list of abbreviated URL IDs to delete
func (h *Handler) DeleteUserLinksBatch() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, err := middleware.TakeUserID(r.Context())
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		deleteURLs := make([]string, 0)
		if err := json.NewDecoder(r.Body).Decode(&deleteURLs); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		err = h.services.Users.DeleteBatchURL(r.Context(), userID, deleteURLs)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		w.WriteHeader(http.StatusAccepted)
	}
}
