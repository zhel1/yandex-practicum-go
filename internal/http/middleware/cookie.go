// Package middleware provides various middleware functionality.
package middleware

import (
	"context"
	"errors"
	"fmt"
	"github.com/zhel1/yandex-practicum-go/internal/dto"
	"github.com/zhel1/yandex-practicum-go/internal/service"
	"net/http"

	"github.com/google/uuid"
)

type CookieHandler struct {
	services *service.Services
}

func NewCookieHandler(services *service.Services) *CookieHandler {
	if services == nil {
		panic(fmt.Errorf("nil services was passed to service URL Handler initializer"))
	}

	return &CookieHandler{
		services: services,
	}
}

func (h *CookieHandler) CookieHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userIDCookie, err := r.Cookie(dto.UserIDCtxName.String())

		var userID string
		if errors.Is(err, http.ErrNoCookie) { //no cookie
			userID = uuid.New().String()
			http.SetCookie(w, h.CreateNewCookie(r.Context(), userID))
		} else if err != nil {
			http.Error(w, "Cookie crumbled", http.StatusInternalServerError)
		} else { //cookie found
			userID, err = h.services.Users.CheckToken(r.Context(), userIDCookie.Value)
			if err != nil {
				http.SetCookie(w, h.CreateNewCookie(r.Context(), userID))
			}
		}

		userIDCtxName := dto.UserIDCtxName
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), userIDCtxName, userID)))
	})
}

func TakeUserID(context context.Context) (string, error) {
	userIDCtx := ""
	if id := context.Value(dto.UserIDCtxName); id != nil {
		userIDCtx = id.(string)
	}

	if userIDCtx == "" {
		return "", errors.New("empty user id")
	}
	return userIDCtx, nil
}

//**********************************************************************************************************************

func (h *CookieHandler) CreateNewCookie(ctx context.Context, userID string) *http.Cookie {
	token, err := h.services.Users.CreateNewToken(ctx, userID)
	if err != nil {
		panic(err.Error())
	}

	cookie := &http.Cookie{
		Name:  dto.UserIDCtxName.String(),
		Value: token,
		Path:  "/",
	}
	return cookie
}
