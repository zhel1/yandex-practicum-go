package middleware

import (
	"context"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/zhel1/yandex-practicum-go/internal/utils"
	"net/http"
)

const UserIDCtxName = "UserID"

type CookieHandler struct {
	cr *utils.Crypto
}

func NewCookieHandler(cr *utils.Crypto) (*CookieHandler, error) {
	if cr == nil {
		return nil, fmt.Errorf("nil Storage was passed to service URL Handler initializer")
	}
	return &CookieHandler{
		cr: cr,
	}, nil
}

func (h *CookieHandler)CokieHandle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userIDCookie, err := r.Cookie(UserIDCtxName)

		var cookieUserID string
		if errors.Is(err, http.ErrNoCookie) { //no cookie
			http.SetCookie(w, h.CreateNewCookie(&cookieUserID))
			//r.AddCookie(newCookie) //TODO delete
		} else if err != nil {
			http.Error(w, "Cookie crumbled", http.StatusInternalServerError)
		} else { //cookie found
			fmt.Println("cookie found")

			cookieUserID, err = h.cr.Decode(userIDCookie.Value)
			if err != nil {
				http.SetCookie(w, h.CreateNewCookie(&cookieUserID))
			}
		}

		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), UserIDCtxName, cookieUserID)))
	})
}
//**********************************************************************************************************************
func (h *CookieHandler)CreateNewCookie(userID *string) *http.Cookie {
	*userID = uuid.New().String()
	token := h.cr.Encode(*userID)
	cookie := &http.Cookie{
		Name:  UserIDCtxName,
		Value: token,
		Path:  "/",
	}
	return cookie
}