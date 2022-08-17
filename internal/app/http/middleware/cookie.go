package middleware

import (
	"context"
	"errors"
	"fmt"
	"github.com/zhel1/yandex-practicum-go/internal/app/dto"
	"net/http"

	"github.com/google/uuid"
	"github.com/zhel1/yandex-practicum-go/internal/app/utils"
)

type CookieHandler struct {
	cr *utils.Crypto
}

func NewCookieHandler(cr *utils.Crypto) *CookieHandler {
	if cr == nil {
		panic(fmt.Errorf("nil Storage was passed to service URL Handler initializer"))
	}

	return &CookieHandler{
		cr: cr,
	}
}

func (h *CookieHandler) CookieHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userIDCookie, err := r.Cookie(dto.UserIDCtxName.String())

		var cookieUserID string
		if errors.Is(err, http.ErrNoCookie) { //no cookie
			http.SetCookie(w, h.CreateNewCookie(&cookieUserID))
			//r.AddCookie(newCookie) //TODO delete
		} else if err != nil {
			http.Error(w, "Cookie crumbled", http.StatusInternalServerError)
		} else { //cookie found
			cookieUserID, err = h.cr.Decode(userIDCookie.Value)
			if err != nil {
				http.SetCookie(w, h.CreateNewCookie(&cookieUserID))
			}
		}

		userIDCtxName := dto.UserIDCtxName
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), userIDCtxName, cookieUserID)))
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
func (h *CookieHandler) CreateNewCookie(userID *string) *http.Cookie {
	*userID = uuid.New().String()
	token := h.cr.Encode(*userID)
	cookie := &http.Cookie{
		Name:  dto.UserIDCtxName.String(),
		Value: token,
		Path:  "/",
	}
	return cookie
}
