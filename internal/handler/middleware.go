package handler

import (
	"context"
	"errors"
	"messages/internal/apperr"
	"net/http"
)

type ctxKey string

const userIDKey ctxKey = "userID"

func UserIDFromContext(ctx context.Context) (int, bool) {
	id, ok := ctx.Value(userIDKey).(int)
	return id, ok // !ok only if code error
}

func (h *AuthHandler) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("session")
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		hashed := hashToken(cookie.Value)
		userID, err := h.sessionRepo.GetUserIDByToken(r.Context(), hashed)
		if err != nil {
			if errors.Is(err, apperr.ErrSessionNotFound) {
				http.Error(w, err.Error(), http.StatusUnauthorized)
			} else {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			return
		}

		ctx := context.WithValue(r.Context(), userIDKey, userID)

		next.ServeHTTP(w, r.WithContext(ctx))

	})
}
